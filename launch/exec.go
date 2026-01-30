package launch

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common/appcfg"
	model2 "github.com/vela-ssoc/ssoc-common/datalayer/model"
	query2 "github.com/vela-ssoc/ssoc-common/datalayer/query"
	"github.com/vela-ssoc/ssoc-common/logger"
	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-common/preadtls"
	"github.com/vela-ssoc/ssoc-common/shipx"
	"github.com/vela-ssoc/ssoc-common/sqldb"
	"github.com/vela-ssoc/ssoc-common/tlscert"
	"github.com/vela-ssoc/ssoc-common/validation"
	brkrestapi "github.com/vela-ssoc/ssoc-manager/application/broker/restapi"
	brkservice "github.com/vela-ssoc/ssoc-manager/application/broker/service"
	"github.com/vela-ssoc/ssoc-manager/application/expose/restapi"
	"github.com/vela-ssoc/ssoc-manager/config"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/muxaccept"
	"github.com/xgfone/ship/v5"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Run2 通过本地配置文件启动程序。
func Run2(ctx context.Context, cfgPath string) error {
	const maxConfigFileSize = 1 << 16
	cfr := appcfg.NewJSON[config.Config](cfgPath, maxConfigFileSize)

	return Exec(ctx, cfr)
}

// Exec 通过配置加载器启动程序。
//
//goland:noinspection GoUnhandledErrorResult
func Exec(ctx context.Context, cfr appcfg.Reader[config.Config]) error {
	logOpts := &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}
	logh := logger.NewMultiHandler(logger.NewTint(os.Stdout, logOpts))
	log := slog.New(logh)

	valid := validation.New()
	if err := valid.RegisterCustomValidations(validation.All()); err != nil {
		log.Error("注册自定义校验器错误", "error", err)
		return err
	}

	cfg, err := cfr.Read(ctx)
	if err != nil {
		log.Error("加载配置文件错误", "error", err)
		return err
	}
	if err = valid.Validate(cfg); err != nil {
		return err
	}

	// 初始化日志。
	logLevel := cfg.Logger.Level
	if logLevel == nil {
		logLevel = new(slog.LevelVar)
	}
	logOpts.Level = logLevel
	logh.Replace()
	if out := cfg.Logger.Lumber(); out != nil {
		defer out.Close()
		h := slog.NewJSONHandler(out, logOpts)
		logh.Append(h)
	}
	if cfg.Logger.Console {
		h := logger.NewTint(os.Stdout, logOpts)
		logh.Append(h)
	}
	log.Info("日志组件初始化完毕")

	log.Info("开始连接接数据库")
	dbCfg := cfg.Database

	gormLogCfg := gormlogger.Config{LogLevel: gormlogger.Info}
	gormLog := logger.NewGorm(logh, gormLogCfg)
	gormCfg := &gorm.Config{Logger: gormLog}
	db, err := sqldb.Open(dbCfg.DSN, gormCfg)
	if err != nil {
		log.Error("数据库连接错误", "error", err)
		return err
	}
	dialectName := db.Dialector.Name()
	log.Info("数据库连接成功", "dialect", dialectName)

	if err = db.Migrator().AutoMigrate(model.All()...); err != nil {
		log.Error("合并数据库差异出错", "error", err)
		return err
	}
	if err = db.Migrator().AutoMigrate(model2.All()...); err != nil {
		log.Error("合并数据库差异出错", "error", err)
		return err
	}
	log.Info("数据库差异合并结束")
	qry := query.Use(db)
	qry2 := query2.Use(db)

	shipLog := shipx.NewLog(logh)
	tunHandler := ship.Default()
	httpHandler := ship.Default()
	httpsHandler := ship.Default()
	tunHandler.NotFound = shipx.NotFound
	httpHandler.NotFound = shipx.NotFound
	httpsHandler.NotFound = shipx.NotFound
	tunHandler.HandleError = shipx.HandleError
	httpHandler.HandleError = shipx.HandleError
	httpsHandler.HandleError = shipx.HandleError
	tunHandler.Logger = shipLog
	httpHandler.Logger = shipLog
	httpsHandler.Logger = shipLog
	tunHandler.Validator = valid
	httpHandler.Validator = valid
	httpsHandler.Validator = valid

	hub := muxserver.NewBrokerHub()
	brkHubHookSvc := brkservice.NewHubHook(log)
	brkBootConfigSvc := brkservice.NewBootConfig(dbCfg)
	brkBrokerStatusSvc := brkservice.NewBrokerStatus(qry, log)

	acceptOpt := muxaccept.Options{
		Huber:      hub,
		Handler:    tunHandler,
		Validator:  valid.Validate,
		Logger:     log,
		BootLoader: brkBootConfigSvc,
		Notifier:   brkHubHookSvc,
	}
	accept := muxaccept.NewAccept(qry2, acceptOpt)

	tunHeartbeatSvc := brkservice.NewHeartbeat(qry, log)
	tunAPIs := []shipx.RouteBinder{ // tunnel API，用于 manager-broker 内部业务通信。
		brkrestapi.NewHeartbeat(tunHeartbeatSvc),
		brkrestapi.NewChat(),
	}
	httpAPIs := []shipx.RouteBinder{ // HTTPS 端口兼容 HTTP 请求，仅用于部署逻辑的暴露。

	}
	httpsAPIs := []shipx.RouteBinder{ // https 主业务
		restapi.NewTunnel(accept),
	}
	{
		rbg := tunHandler.Group("/api/v1")
		if err = shipx.BindRoutes(rbg, tunAPIs); err != nil {
			log.Error("注册 tunnel 路由出错", "error", err)
			return err
		}
	}
	{
		rbg := httpHandler.Group("/api/v1")
		if err = shipx.BindRoutes(rbg, httpAPIs); err != nil {
			log.Error("注册 http 路由出错", "error", err)
			return err
		}
	}
	{
		rbg := httpsHandler.Group("/api/v1")
		if err = shipx.BindRoutes(rbg, httpsAPIs); err != nil {
			log.Error("注册 https 路由出错", "error", err)
			return err
		}
	}

	noopTLS := func(context.Context) ([]*tls.Certificate, error) { return nil, nil }
	tlsPool := tlscert.NewMatch(noopTLS, log)
	httpSrv := &http.Server{
		Handler: httpHandler,
	}
	httpsSrv := &http.Server{
		Handler: httpsHandler,
		TLSConfig: &tls.Config{
			GetCertificate: tlsPool.GetCertificate,
		},
	}

	srvCfg := cfg.Server
	addr := srvCfg.Addr
	if addr == "" {
		addr = ":443"
	}
	lis, err := preadtls.ListenTCP(addr, 5*time.Second)
	if err != nil {
		log.Error("网络监听失败", "error", err)
		return err
	}
	defer lis.Close()

	_ = brkBrokerStatusSvc.Reset(ctx)
	log.Info("启动 http https 服务")
	errs := make(chan error, 1)
	go listenHTTP(errs, httpSrv, lis.TCPListener())
	go listenHTTPS(errs, httpsSrv, lis.TLSListener())

	select {
	case <-ctx.Done():
		err = ctx.Err()
	case err = <-errs:
	}

	_ = httpSrv.Close()
	_ = httpsSrv.Close()
	_ = brkBrokerStatusSvc.Reset(ctx)

	return err
}

func listenHTTP(errs chan<- error, srv *http.Server, ln net.Listener) {
	errs <- srv.Serve(ln)
}

func listenHTTPS(errs chan<- error, srv *http.Server, ln net.Listener) {
	errs <- srv.ServeTLS(ln, "", "")
}
