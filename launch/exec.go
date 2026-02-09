package launch

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/vela-ssoc/ssoc-common/appcfg"
	"github.com/vela-ssoc/ssoc-common/cronv3"
	"github.com/vela-ssoc/ssoc-common/logger"
	"github.com/vela-ssoc/ssoc-common/mongodb"
	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-common/preadtls"
	"github.com/vela-ssoc/ssoc-common/shipx"
	"github.com/vela-ssoc/ssoc-common/store/repository"
	"github.com/vela-ssoc/ssoc-common/tlscert"
	"github.com/vela-ssoc/ssoc-common/validation"
	"github.com/vela-ssoc/ssoc-common/vmetric"
	brkrestapi "github.com/vela-ssoc/ssoc-manager/application/broker/restapi"
	brkservice "github.com/vela-ssoc/ssoc-manager/application/broker/service"
	"github.com/vela-ssoc/ssoc-manager/application/current/cronjob"
	curservice "github.com/vela-ssoc/ssoc-manager/application/current/service"
	"github.com/vela-ssoc/ssoc-manager/application/expose/restapi"
	expservice "github.com/vela-ssoc/ssoc-manager/application/expose/service"
	"github.com/vela-ssoc/ssoc-manager/config"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/brkclient"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/muxaccept"
	"github.com/vela-ssoc/ssoc-proto/muxtool"
	"github.com/xgfone/ship/v5"
)

func Exec(ctx context.Context, cfg string) error {
	cfr := appcfg.NewJSON[config.Config](cfg)

	return Run(ctx, cfr)
}

func Run(ctx context.Context, cfr appcfg.Reader[config.Config]) error {
	cfg, err := cfr.Read(ctx)
	if err != nil {
		return err
	}

	return Start(ctx, cfg)
}

//goland:noinspection GoUnhandledErrorResult
func Start(ctx context.Context, cfg *config.Config) error {
	// 项目启动时还未连接到中心端，此时要默认一个日志输出。
	logh := initLogHandler()
	log := slog.New(logh)
	log.Info("初始日志组件装配完毕")

	valid := validation.New()
	if err := valid.RegisterCustomValidations(validation.All()); err != nil {
		log.Error("校验器注册出错", "error", err)
		return err
	}

	if err := valid.Validate(cfg); err != nil {
		log.Error("配置校验出错", "error", err)
		return err
	}
	log.Info("配置加载成功")

	logh.Replace() // 清空默认日志，改为配置文件设置的日志输出方式。
	srvCfg, dbCfg, logCfg := cfg.Server, cfg.Database, cfg.Logger
	logLevel := logCfg.Level.LevelVar()
	logOpts := &slog.HandlerOptions{Level: logLevel, AddSource: true}
	if logCfg.Console {
		h := logger.NewTint(os.Stdout, logOpts)
		logh.Append(h)
	}
	if file := logCfg.Lumber(); file != nil {
		defer file.Close()
		h := slog.NewJSONHandler(file, logOpts)
		logh.Append(h)
	}
	log.Info("日志初始化完毕")

	log.Info("开始连接接数据库")
	mdb, err := mongodb.Connect(dbCfg.URI)
	if err != nil {
		log.Error("数据库连接错误", "error", err)
		return err
	}
	log.Info("数据库连接成功，开始初始化索引")
	db := repository.NewDB(mdb, log)
	if err = db.CreateIndex(ctx); err != nil {
		log.Error("数据库索引初始化错误", "error", err)
		return err
	}

	crontab := cronv3.New(log)
	crontab.Start()

	shipLog := shipx.NewLog(logh)
	httpsSH := ship.Default()
	httpSH := ship.Default()
	brokSH := ship.Default()
	httpsSH.NotFound = shipx.NotFound
	httpSH.NotFound = shipx.NotFound
	brokSH.NotFound = shipx.NotFound
	httpsSH.HandleError = shipx.HandleError
	httpSH.HandleError = shipx.HandleError
	brokSH.HandleError = shipx.HandleError
	httpsSH.Logger = shipLog
	httpSH.Logger = shipLog
	brokSH.Logger = shipLog
	httpsSH.Validator = valid
	httpSH.Validator = valid
	brokSH.Validator = valid

	hub := muxserver.NewBrokerHub()
	brkHubHookSvc := brkservice.NewHubHook(log)
	curBrokerSvc := curservice.NewBroker(db, dbCfg, log)
	expBrokerSvc := expservice.NewBroker(db, log)
	expPyroscopeSvc := expservice.NewPyroscopeConfig(db, log)
	expVictoriaMetricsSvc := expservice.NewVictoriaMetricsConfig(db, log)

	sysdial := new(net.Dialer)
	mixdial := muxserver.NewMixedDialer(nil, hub, sysdial)
	basecli := muxtool.NewClient(mixdial, log)
	brkcli := brkclient.NewClient(basecli)

	acceptOpt := muxaccept.Options{
		Huber:      hub,
		Handler:    brokSH,
		Validator:  valid.Validate,
		Logger:     log,
		BootLoader: curBrokerSvc,
		Notifier:   brkHubHookSvc,
	}
	accept := muxaccept.NewAccept(db, acceptOpt)

	brkHeartbeatSvc := brkservice.NewHeartbeat(db, log)
	brokAPIs := []shipx.RouteBinder{ // tunnel API，用于 manager-broker 内部业务通信。
		brkrestapi.NewHeartbeat(brkHeartbeatSvc),
	}
	httpAPIs := []shipx.RouteBinder{ // HTTPS 端口兼容 HTTP 请求，仅用于部署逻辑的暴露。

	}
	httpsAPIs := []shipx.RouteBinder{ // https 主业务
		restapi.NewBroker(expBrokerSvc),
		restapi.NewBrokerTunnel(brkcli),
		restapi.NewTunnel(accept),
	}

	for k, v := range srvCfg.Static { // 注册静态资源
		httpsSH.Route(k).Static(v)
	}
	{
		routes := append(httpsAPIs, httpAPIs...)
		rbg := httpsSH.Group("/api/v1")
		if err = shipx.BindRoutes(rbg, routes); err != nil {
			log.Error("注册 https 路由出错", "error", err)
			return err
		}
	}
	{
		rbg := httpSH.Group("/api/v1")
		if err = shipx.BindRoutes(rbg, httpAPIs); err != nil {
			log.Error("注册 http 路由出错", "error", err)
			return err
		}
	}
	{
		rbg := brokSH.Group("/api/v1")
		if err = shipx.BindRoutes(rbg, brokAPIs); err != nil {
			log.Error("注册 broker 路由出错 (internal)", "error", err)
			return err
		}
	}

	managerLabel := vmetric.ManagerLabel()
	metricWriters := []vmetric.MetricWriter{vmetric.NewPsutil()}
	cronTasks := []cronv3.Tasker{
		cronjob.NewMetrics(managerLabel, expVictoriaMetricsSvc, metricWriters),
	}
	if err = crontab.AddTasks(cronTasks); err != nil {
		log.Error("定时任务注册出错", "error", err)
		return err
	}

	crtFile := tlscert.NewFile("resources/tls/server.pem", "resources/tls/server.key")
	crtPool := tlscert.NewMatch(crtFile, log)
	httpSrv := &http.Server{Handler: httpSH}
	httpsSrv := &http.Server{Handler: httpsSH, TLSConfig: &tls.Config{GetCertificate: crtPool.GetCertificate}}

	listenAddr := srvCfg.Addr
	if listenAddr == "" {
		listenAddr = ":443"
	}
	lis, err := preadtls.ListenTCP(listenAddr, 10*time.Second)
	if err != nil {
		log.Error("服务监听出错", "error", err)
		return err
	}
	defer lis.Close()

	// 重置
	if err1 := curBrokerSvc.Reset(time.Minute); err1 != nil {
		log.Warn("重置 broker 在线状态出错", "error", err1)
	}
	if err1 := expPyroscopeSvc.Start(ctx); err1 != nil {
		log.Warn("pyroscope 上报程序启动出错", "error", err1)
	}

	errs := make(chan error, 1)
	go serveHTTP(errs, httpSrv, lis.TCPListener())
	go serveHTTPS(errs, httpsSrv, lis.TLSListener())

	select {
	case <-ctx.Done():
	case err = <-errs:
	}

	crontab.Stop()
	_ = httpSrv.Close()
	_ = httpsSrv.Close()
	curBrokerSvc.Reset(10 * time.Second)

	cause := context.Cause(ctx)
	log.Error("程序停止运行", "error", err, "cause", cause)

	return err
}

func serveHTTP(errs chan<- error, srv *http.Server, ln net.Listener) {
	errs <- srv.Serve(ln)
}

func serveHTTPS(errs chan<- error, srv *http.Server, ln net.Listener) {
	errs <- srv.ServeTLS(ln, "", "")
}
