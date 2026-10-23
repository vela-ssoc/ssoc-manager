package launch

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/NVIDIA/gontainer/v2"
	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/varload"
	"github.com/vela-ssoc/ssoc-common-mb/dbconn"
	"github.com/vela-ssoc/ssoc-common-mb/logger"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/ssoc-common-mb/tlscert"
	"github.com/vela-ssoc/ssoc-common-mb/validation"
	"github.com/vela-ssoc/ssoc-manager/appv3/restapi"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/vela-ssoc/ssoc-manager/config"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/muxcentral"
	"github.com/xgfone/ship/v5"
)

func RunV3(ctx context.Context, filename string) error {
	var bootcfg config.Config
	vl := varload.JSONC(filename)
	err := vl.Load(ctx, &bootcfg)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		bootcfg, err = runOOBE(ctx, filename)
	}
	if err != nil {
		return err
	}

	return Exec(ctx, bootcfg)
}

//goland:noinspection GoUnhandledErrorResult
func Exec(ctx context.Context, bootcfg config.Config) error {
	log, logc := logger.NewInit(config.DefaultLogFilename)
	defer logc.Close()

	valid := validation.New()
	if err := valid.RegisterCustomValidations(validation.All()); err != nil {
		log.Error("注册扩展校验器出错", "err", err)
		return err
	}
	if err := valid.Validate(bootcfg); err != nil {
		log.Error("校验启动配置文件不通过", "err", err)
		return err
	}

	dbcfg := bootcfg.Database
	db, err := dbconn.OpenMDB(dbcfg.URI)
	if err != nil {
		log.Error("连接 MongoDB 出错", "err", err)
		return err
	}
	defer closeMDB(db)

	mdb := repository.NewBaseDB(db, log)
	if err = mdb.CreateIndex(ctx); err != nil {
		log.Error("创建 MongoDB 索引出错", "err", err)
		return err
	}
	cfg, err := mdb.ManagerConfig().Get(ctx)
	if err != nil {
		log.Error("加载全量配置文件出错", "err", err)
		return err
	}
	if err = valid.Validate(cfg); err != nil {
		log.Error("加载全量配置文件校验不通过", "err", err)
		return err
	}
	log.Info("全量配置文件加载并校验成功")

	// 替换初始日志
	var handlers []slog.Handler
	{
		logcfg := cfg.Logger
		level := new(slog.LevelVar)
		_ = level.UnmarshalText([]byte(logcfg.Level))
		opts := &slog.HandlerOptions{AddSource: true, Level: level}

		if logcfg.Console {
			handlers = append(handlers, logger.NewTint(os.Stdout, opts))
		}
		if lumber := logcfg.Lumber(); lumber != nil {
			defer lumber.Close()
			handlers = append(handlers, slog.NewJSONHandler(lumber, opts))
		}
	}
	logHandler := slog.NewMultiHandler(handlers...)
	log = slog.New(logHandler)
	_ = logc.Close()

	rdb, err := dbconn.OpenRDB(dbcfg.DSN)
	if err != nil {
		log.Error("连接关系型数据库出错", "err", err)
		return err
	}
	defer closeRDB(rdb)

	log.Info("关系数据库连接成功", "dialect", rdb.Dialector.Name())
	// 如果需要创建关系型数据库的索引、数据合并，在此处实现。
	qry := query.Use(rdb)
	pdb := polydb.NewSource(mdb, qry) // 数据库聚合

	// 系统拨号器
	sysdial := net.Dialer{
		Timeout: 30 * time.Second,
		KeepAliveConfig: net.KeepAliveConfig{
			Enable:   true,
			Idle:     30 * time.Second,
			Interval: 15 * time.Second,
			Count:    3,
		},
	}
	sysdial.SetMultipathTCP(true)
	hub := muxproto.NewBHub()
	dialer := muxproto.NewDualDialer(hub, sysdial)
	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 2,
		MaxConnsPerHost:     20,
		IdleConnTimeout:     30 * time.Second,
	}
	httpcli := &http.Client{Transport: transport}
	tlspool := tlscert.NewMatch(nil, log)

	exposeMUX := newShip(valid, logHandler)
	brokerMUX := newShip(valid, logHandler)

	centralopts := muxcentral.Options{
		Validate:   valid.Validate,
		Handler:    brokerMUX,
		Logger:     log,
		PerTimeout: 10 * time.Second,
	}
	wsupgrade := &websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		// CheckOrigin:       nil,
		EnableCompression: true,
	}

	route := shipx.RouteGroup3{
		ExposeV1: shipx.RouteEndpoint{
			Path:  "/api/v1",
			Group: exposeMUX.Group("/api/v1"),
		},
		ExposeV3: shipx.RouteEndpoint{
			Path:  "/api/v3",
			Group: exposeMUX.Group("/api/v3"),
		},
		BrokerV1: shipx.RouteEndpoint{
			Path:  "/api/v1",
			Group: brokerMUX.Group("/api/v1"),
		},
		BrokerV3: shipx.RouteEndpoint{
			Path:  "/api/v3",
			Group: brokerMUX.Group("/api/v3"),
		},
	}

	diopts := []gontainer.Option{
		gontainer.NewService(bootcfg),
		gontainer.NewService(cfg),
		gontainer.NewService(log),
		gontainer.NewService(qry),
		gontainer.NewService(pdb),
		gontainer.NewService(hub),
		gontainer.NewService(route),
		gontainer.NewService(centralopts),
		gontainer.NewService(httpcli),
		gontainer.NewService(transport),
		gontainer.NewService(wsupgrade),

		gontainer.NewFactory(muxcentral.NewServer),
		gontainer.NewFactory(service.NewBroker),
		gontainer.NewFactory(service.NewBrokerConnectHistory),
		gontainer.NewFactory(service.NewBrokerStream),
		gontainer.NewFactory(service.NewReverseProxy),
		gontainer.NewFactory(service.NewZipArchive),
		gontainer.NewFactory(service.NewZipFile),
		gontainer.NewFactory(restapi.NewBroker),
		gontainer.NewFactory(restapi.NewBrokerConnectHistory),
		gontainer.NewFactory(restapi.NewBrokerReverseProxy),
		gontainer.NewFactory(restapi.NewBrokerStream),
		gontainer.NewFactory(restapi.NewConfig),
		gontainer.NewFactory(restapi.NewReverseProxy),
		gontainer.NewFactory(restapi.NewSystemPTY),
		gontainer.NewFactory(restapi.NewZipArchive),
		gontainer.NewFactory(restapi.NewTunnel),
		gontainer.NewService(restapi.NewWebDAV("/")),
		gontainer.NewFactory(restapi.NewZipFile),

		gontainer.NewEntrypoint(registerRoutes),
	}
	if err = gontainer.Run(diopts...); err != nil {
		log.Error("依赖注入出错", "err", err)
		fmt.Println(err.Error())
		return err
	}
	log.Info("依赖注入执行完毕")

	srv := &http.Server{
		Addr:      cfg.Server.Addr,
		Handler:   exposeMUX,
		TLSConfig: &tls.Config{GetCertificate: tlspool.GetCertificate},
	}
	errch := make(chan error, 1)
	go startHTTPS(errch, srv)
	select {
	case <-ctx.Done():
		err = context.Cause(ctx)
	case err = <-errch:
	}

	log.Error("程序运行结束", "err", err)
	_ = closeHTTP(srv)

	return err
}

func newShip(v *validation.Validate, h slog.Handler) *ship.Ship {
	log := logger.NewFormat(h, 4)

	s := ship.Default()
	s.Validator = v
	s.Logger = log

	return s
}
