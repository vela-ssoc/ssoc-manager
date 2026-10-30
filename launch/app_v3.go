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
	"github.com/robfig/cron/v3"
	"github.com/vela-ssoc/ssoc-common-mb/crontab"
	"github.com/vela-ssoc/ssoc-common-mb/dal/gridfs"
	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/varload"
	"github.com/vela-ssoc/ssoc-common-mb/dbconn"
	"github.com/vela-ssoc/ssoc-common-mb/gopool"
	"github.com/vela-ssoc/ssoc-common-mb/httpx"
	"github.com/vela-ssoc/ssoc-common-mb/integration/cmdb"
	"github.com/vela-ssoc/ssoc-common-mb/integration/elastic"
	"github.com/vela-ssoc/ssoc-common-mb/logger"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/ssoc-common-mb/profile"
	"github.com/vela-ssoc/ssoc-common-mb/pubsub"
	"github.com/vela-ssoc/ssoc-common-mb/storage/v2"
	"github.com/vela-ssoc/ssoc-common-mb/tlscert"
	"github.com/vela-ssoc/ssoc-common-mb/validation"
	"github.com/vela-ssoc/ssoc-common-mb/vmexport"
	"github.com/vela-ssoc/ssoc-manager/app/mgtapi"
	mgtsvc "github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/middle"
	"github.com/vela-ssoc/ssoc-manager/appv3/restapi"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/vela-ssoc/ssoc-manager/bridge/linkhub"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/config"
	"github.com/vela-ssoc/ssoc-manager/cronfunc"
	"github.com/vela-ssoc/ssoc-manager/integration/ywcmdb"
	"github.com/vela-ssoc/ssoc-manager/migration"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/muxcentral"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/muxroute"
	"github.com/vela-ssoc/vela-common-mba/netutil"
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
		opts := &slog.HandlerOptions{AddSource: true, Level: logcfg.LevelVar()}
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
	if err = rdb.WithContext(ctx).AutoMigrate(model.All()...); err != nil {
		log.Error("关系数据库执行迁移出错", "err", err)
		return err
	}
	log.Info("关系数据库迁移完毕")
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

	muxresolvers := []muxproto.TLDResolver{
		muxproto.NewHubResolver(hub),           // 通往中继节点
		muxroute.NewAgentRouter(pdb, hub, log), // 通往探针节点
	}
	dialer := muxproto.NewTLDDialer(sysdial, muxresolvers)
	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConns:        20,
		MaxIdleConnsPerHost: 2,
		MaxConnsPerHost:     20,
		IdleConnTimeout:     30 * time.Second,
	}
	httpcli := &http.Client{Transport: transport}
	httpxcli := httpx.NewClient(httpcli)
	tlspool := tlscert.NewMatch(nil, log)
	pool := gopool.New(4096)

	cronparser := crontab.NewSecondOptional()
	crond := crontab.NewTable(log, cron.WithParser(cronparser))
	crond.Start()
	defer crond.Stop()

	exposeMUX := newShip(valid, logHandler)
	brokerMUX := newShip(valid, logHandler)
	cfgsrv := cfg.Server
	for k, v := range cfgsrv.Static {
		exposeMUX.Route(k).StaticFS(v)
	}

	centralopts := muxcentral.Options{
		Validate:   valid.Validate,
		Handler:    brokerMUX,
		Logger:     log,
		PerTimeout: 10 * time.Second,
	}
	wsupgrade := &websocket.Upgrader{
		HandshakeTimeout:  10 * time.Second,
		EnableCompression: true,
	}

	if err = migration.Exec(ctx, migration.NewUser(pdb, log)); err != nil {
		log.Error("数据迁移出错", "err", err)
		return err
	}

	const apiV1, apiV3 = "/api/v1", "/api/v3"
	noauth := exposeMUX.Group(apiV1)

	midauth := middle.NewAuth(pdb, 24*time.Hour, log)
	v1auth := noauth.Clone().Use(midauth)

	route := shipx.RouteGroup3{
		ExposeV1: shipx.RouteEndpoint{
			Path:  apiV1,
			Group: v1auth,
		},
		ExposeV3: shipx.RouteEndpoint{
			Path:  apiV3,
			Group: exposeMUX.Group(apiV3).Use(midauth),
		},
		BrokerV1: shipx.RouteEndpoint{
			Path:  apiV1,
			Group: brokerMUX.Group(apiV1),
		},
		BrokerV3: shipx.RouteEndpoint{
			Path:  apiV3,
			Group: brokerMUX.Group(apiV3),
		},
		ExposeV1NoAuth: shipx.RouteEndpoint{
			Path:  apiV1,
			Group: noauth,
		},
	}
	dispatch := pubsub.NewDispatcher()

	v1cfg := new(profile.ManagerConfig)
	hubv1 := linkhub.New(qry, brokerMUX, pool, v1cfg) // v1 fake hub
	pushv1 := push.NewPush(qry, hubv1)
	diopts := []gontainer.Option{
		gontainer.NewService(bootcfg),
		gontainer.NewService(cfg),
		gontainer.NewService(log),
		gontainer.NewService(db),
		gontainer.NewService(qry),
		gontainer.NewService(pdb),
		gontainer.NewService(hub),
		gontainer.NewService(route),
		gontainer.NewService(centralopts),
		gontainer.NewService(httpcli),
		gontainer.NewService(httpxcli),            // v1
		gontainer.NewService(hubv1),               // v1
		gontainer.NewService(pushv1),              // v1
		gontainer.NewService(pool),                // v1
		gontainer.NewService(netutil.NewClient()), // v1
		gontainer.NewService(transport),
		gontainer.NewService(wsupgrade),
		gontainer.NewService(dispatch),

		// --------------------[ V1 START]--------------------
		gontainer.NewFactory(cmdb.NewConfigure),
		gontainer.NewFactory(cmdb.NewClient),
		gontainer.NewService(crond),
		gontainer.NewService(elastic.NewConfigure(qry, "ssoc")),
		gontainer.NewFactory(elastic.NewSearch),

		gontainer.NewFactory(gridfs.NewFS),
		gontainer.NewFactory(storage.NewStore),
		gontainer.NewFactory(mgtsvc.NewDigest),
		gontainer.NewFactory(mgtsvc.NewMinionFilter),
		gontainer.NewFactory(mgtsvc.NewSequence),
		gontainer.NewService(mgtsvc.NewLoginLock(qry, time.Hour, 10)),
		gontainer.NewFactory(mgtapi.NewAccount),
		gontainer.NewFactory(mgtsvc.NewAccount),
		gontainer.NewFactory(mgtsvc.NewAlertServer),
		gontainer.NewFactory(mgtapi.NewAlertServer),
		gontainer.NewFactory(mgtapi.NewAuth),
		gontainer.NewFactory(mgtsvc.NewAuth),
		gontainer.NewFactory(mgtsvc.NewBroker),
		gontainer.NewFactory(mgtapi.NewBroker),
		gontainer.NewFactory(mgtsvc.NewBrokerBinary),
		gontainer.NewFactory(mgtapi.NewBrokerBinary),
		gontainer.NewFactory(mgtsvc.NewCert),
		gontainer.NewFactory(mgtapi.NewCert),
		gontainer.NewFactory(mgtsvc.NewDash),
		gontainer.NewFactory(mgtapi.NewDash),
		gontainer.NewFactory(mgtsvc.NewMinion),
		gontainer.NewFactory(mgtapi.NewMinion),
		gontainer.NewFactory(mgtsvc.NewMinionAccount),
		gontainer.NewFactory(mgtapi.NewMinionAccount),
		gontainer.NewFactory(mgtsvc.NewMinionBinary),
		gontainer.NewFactory(mgtapi.NewMinionBinary),
		gontainer.NewFactory(mgtsvc.NewMinionCustomized),
		gontainer.NewFactory(mgtapi.NewMinionCustomized),
		gontainer.NewFactory(mgtsvc.NewMinionListen),
		gontainer.NewFactory(mgtapi.NewMinionListen),
		gontainer.NewFactory(mgtsvc.NewMinionLogon),
		gontainer.NewFactory(mgtapi.NewMinionLogon),
		gontainer.NewFactory(mgtsvc.NewMinionTask),
		gontainer.NewFactory(mgtapi.NewMinionTask),
		gontainer.NewFactory(mgtsvc.NewNotifier),
		gontainer.NewFactory(mgtapi.NewNotifier),
		gontainer.NewFactory(mgtsvc.NewOplog),
		gontainer.NewFactory(mgtapi.NewOplog),
		gontainer.NewFactory(mgtsvc.NewPassDNS),
		gontainer.NewFactory(mgtapi.NewPassDNS),
		gontainer.NewFactory(mgtsvc.NewPassIP),
		gontainer.NewFactory(mgtapi.NewPassIP),
		gontainer.NewFactory(mgtsvc.NewProcess),
		gontainer.NewFactory(mgtapi.NewProcess),
		gontainer.NewFactory(mgtsvc.NewRisk),
		gontainer.NewFactory(mgtapi.NewRisk),
		gontainer.NewFactory(mgtsvc.NewRiskDNS),
		gontainer.NewFactory(mgtapi.NewRiskDNS),
		gontainer.NewFactory(mgtsvc.NewRiskFile),
		gontainer.NewFactory(mgtapi.NewRiskFile),
		gontainer.NewFactory(mgtsvc.NewRiskIP),
		gontainer.NewFactory(mgtapi.NewRiskIP),
		gontainer.NewFactory(mgtsvc.NewSBOMComponent),
		gontainer.NewFactory(mgtapi.NewSBOMComponent),
		gontainer.NewFactory(mgtsvc.NewSBOMProject),
		gontainer.NewFactory(mgtapi.NewSBOMProject),
		gontainer.NewFactory(mgtsvc.NewSBOMVuln),
		gontainer.NewFactory(mgtapi.NewSBOMVuln),
		gontainer.NewFactory(mgtsvc.NewShared),
		gontainer.NewFactory(mgtapi.NewShared),
		gontainer.NewFactory(mgtsvc.NewStartup),
		gontainer.NewFactory(mgtapi.NewStartup),
		gontainer.NewFactory(mgtsvc.NewStore),
		gontainer.NewFactory(mgtapi.NewStore),
		gontainer.NewFactory(mgtsvc.NewSubstance),
		gontainer.NewFactory(mgtapi.NewSubstance),
		gontainer.NewFactory(mgtsvc.NewSubstanceTask),
		gontainer.NewFactory(mgtapi.NewSubstanceTask),
		gontainer.NewFactory(mgtsvc.NewTag),
		gontainer.NewFactory(mgtapi.NewTag),
		gontainer.NewFactory(mgtsvc.NewTaskExecute),
		gontainer.NewFactory(mgtapi.NewTaskExecute),
		gontainer.NewFactory(mgtsvc.NewTaskExecuteItem),
		gontainer.NewFactory(mgtapi.NewTaskExecuteItem),
		gontainer.NewFactory(mgtsvc.NewTaskExtension),
		gontainer.NewFactory(mgtapi.NewTaskExtension),
		gontainer.NewFactory(mgtsvc.NewThird),
		gontainer.NewFactory(mgtapi.NewThird),
		gontainer.NewFactory(mgtsvc.NewThirdCustomized),
		gontainer.NewFactory(mgtapi.NewThirdCustomized),
		gontainer.NewFactory(mgtsvc.NewUser),
		gontainer.NewFactory(mgtapi.NewUser),
		gontainer.NewFactory(mgtsvc.NewVIP),
		gontainer.NewFactory(mgtapi.NewVIP),

		gontainer.NewEntrypoint(registerV1Routes), // 临时补丁方案
		// --------------------[ V1 END]--------------------

		gontainer.NewFactory(muxcentral.NewServer),
		gontainer.NewFactory(restapi.NewAgentReverseProxy),
		gontainer.NewFactory(service.NewBroker),
		gontainer.NewFactory(restapi.NewBroker),
		gontainer.NewFactory(service.NewBrokerConnectHistory),
		gontainer.NewFactory(restapi.NewBrokerConnectHistory),
		gontainer.NewFactory(restapi.NewBrokerReverseProxy),
		gontainer.NewFactory(service.NewBrokerStream),
		gontainer.NewFactory(restapi.NewBrokerStream),
		gontainer.NewFactory(service.NewCMDB),
		gontainer.NewFactory(restapi.NewCMDB),
		gontainer.NewFactory(service.NewCMDBConfig),
		gontainer.NewFactory(restapi.NewCMDBConfig),
		gontainer.NewFactory(restapi.NewItself),
		gontainer.NewFactory(service.NewItselfConfig),
		gontainer.NewFactory(restapi.NewItselfConfig),
		gontainer.NewFactory(restapi.NewItselfCrontab),
		gontainer.NewFactory(restapi.NewItselfExit),
		gontainer.NewFactory(restapi.NewItselfRoute),
		gontainer.NewFactory(service.NewMinion),
		gontainer.NewFactory(restapi.NewMinion),
		gontainer.NewFactory(service.NewOAuth),
		gontainer.NewFactory(restapi.NewOAuth),
		gontainer.NewFactory(service.NewOAuthConfig),
		gontainer.NewFactory(restapi.NewOAuthConfig),
		gontainer.NewFactory(service.NewReverseProxy),
		gontainer.NewFactory(restapi.NewReverseProxy),
		gontainer.NewFactory(restapi.NewSSEEvent),
		gontainer.NewFactory(restapi.NewSystemProcess),
		gontainer.NewFactory(service.NewVictoriaMetrics),
		gontainer.NewFactory(restapi.NewVictoriaMetrics),
		gontainer.NewFactory(restapi.NewSystemPTY),
		gontainer.NewFactory(service.NewZipArchive),
		gontainer.NewFactory(restapi.NewZipArchive),
		gontainer.NewFactory(restapi.NewTunnel),
		gontainer.NewService(restapi.NewWebDAV("/")),
		gontainer.NewFactory(service.NewZipFile),
		gontainer.NewFactory(restapi.NewZipFile),

		gontainer.NewFactory(vmexport.NewMongoDB),
		gontainer.NewFactory(vmexport.NewPsutil),

		gontainer.NewFactory(ywcmdb.NewClient),
		gontainer.NewFactory(cronfunc.NewCMDB),
		gontainer.NewFactory(cronfunc.NewCMDBToSSOC),
		gontainer.NewFactory(cronfunc.NewSSOCToCMDB),
		gontainer.NewFactory(cronfunc.NewVMetric),

		gontainer.NewEntrypoint(cronAddJobs),
		gontainer.NewEntrypoint(registerRoutes),
	}
	if err = gontainer.Run(diopts...); err != nil {
		log.Error("依赖注入出错", "err", err)
		fmt.Println(err.Error())
		return err
	}
	log.Info("依赖注入执行完毕")

	srv := &http.Server{
		Addr:      cfgsrv.Addr,
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
	_ = srv.Close()

	return err
}

func newShip(v *validation.Validate, h slog.Handler) *ship.Ship {
	log := logger.NewFormat(h, 4)

	s := ship.Default()
	s.Validator = v
	s.Logger = log

	return s
}
