package launch

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dal/model"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/sqldb"
	"github.com/vela-ssoc/vela-common-mb/stdlog"
	"github.com/vela-ssoc/vela-common-mb/validate"
	"github.com/vela-ssoc/vela-manager/applet/brkmux"
	brkrestapi "github.com/vela-ssoc/vela-manager/applet/broker/restapi"
	brkservice "github.com/vela-ssoc/vela-manager/applet/broker/service"
	"github.com/vela-ssoc/vela-manager/applet/manager/restapi"
	"github.com/vela-ssoc/vela-manager/applet/manager/service"
	"github.com/vela-ssoc/vela-manager/applet/shipx"
	"github.com/vela-ssoc/vela-manager/profile"
	"github.com/xgfone/ship/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Run1(ctx context.Context, cfg string) error {
	conf, err := profile.JSONC(cfg)
	if err != nil {
		return err
	}

	return Exec(ctx, conf)
}

//goland:noinspection GoUnhandledErrorResult
func Exec(ctx context.Context, cfg *profile.Config) error {
	valid := validate.New()
	if err := valid.Validate(cfg); err != nil { // 校验配置参数。
		return err
	}

	// =====[ 初始化日志组件 ]=====
	logCfg := cfg.Logger
	logWriter := logCfg.Writer()
	defer logWriter.Close() // 程序结束时关闭。

	rawLevel := []byte(logCfg.Level)
	logLevel := new(slog.LevelVar)
	if err := logLevel.UnmarshalText(rawLevel); err != nil {
		logLevel.Set(slog.LevelInfo)
	}
	logOption := &slog.HandlerOptions{AddSource: true, Level: logLevel}
	logHandler := slog.NewJSONHandler(logWriter, logOption)
	log := slog.New(logHandler)
	log.Info("日志组件初始化完毕")

	// =====[ 连接数据库 ]=====
	dbCfg := cfg.Database
	gormLog, gormLevel := sqldb.NewLog(logWriter, logger.Config{LogLevel: cfg.GormLevel()})
	gormCfg := &gorm.Config{Logger: gormLog}
	db, gauss, err := sqldb.Open(dbCfg.DSN, log, gormCfg)
	if err != nil {
		return err
	}
	sdb, err := db.DB()
	if err != nil {
		return err
	}
	defer sdb.Close() // 程序结束时断开数据库连接。

	sdb.SetMaxOpenConns(dbCfg.MaxOpenConn)
	sdb.SetMaxIdleConns(dbCfg.MaxIdleConn)
	sdb.SetConnMaxLifetime(dbCfg.MaxLifeTime.Duration())
	sdb.SetConnMaxIdleTime(dbCfg.MaxIdleTime.Duration())

	if gauss {
		log.Warn("当前连接的是 OpenGauss 信创数据库")
	} else {
		log.Warn("当前连接的是 MySQL 数据库")
	}
	if dbCfg.Migrate { // 迁移合并数据库
		if err = db.AutoMigrate(model.All()...); err != nil {
			return err
		}
	}

	qry := query.Use(db)
	shipLog := shipx.NewLog(slog.New(stdlog.Skip(logHandler, 6)))
	tunMux := ship.Default() // 虚拟通道的 Handler
	webMux := ship.Default() // Web HTTP Handler

	webMux.Validator, tunMux.Validator = valid, valid
	webMux.Logger, tunMux.Logger = shipLog, shipLog
	webMux.NotFound, tunMux.NotFound = shipx.NotFound, shipx.NotFound
	webMux.HandleError, tunMux.HandleError = shipx.HandleError, shipx.HandleError

	brkGateAPI := brkmux.New(qry, tunMux, cfg, log)
	brkLink := brkGateAPI.Linker()

	alertServerSvc := service.NewAlertServer(qry, log)
	brokerSvc := service.NewBroker(brkLink, log)
	logSvc := service.NewLog(logLevel, gormLevel, log)
	webRouters := []shipx.Router{
		restapi.NewAlertServer(alertServerSvc),
		restapi.NewBroker(brokerSvc),
		restapi.NewLog(logSvc),
		restapi.NewSystem(),
		brkGateAPI,
	}
	brkPeerSvc := brkservice.NewPeer(qry, log)
	tunRouters := []shipx.Router{
		brkrestapi.NewPeer(brkPeerSvc),
	}

	tunRGB := tunMux.Group("/")
	if err = shipx.BindRouters(tunRGB, tunRouters); err != nil {
		return err
	}
	webRGB := webMux.Group("/api/v1")
	if err = shipx.BindRouters(webRGB, webRouters); err != nil {
		return err
	}

	srvCfg := cfg.Server
	mgtSrv := &http.Server{Handler: webMux}
	errs := make(chan error, 1)
	go serveHTTP(srvCfg, mgtSrv, errs)
	log.Info("服务准备就绪")
	select {
	case <-ctx.Done():
	case err = <-errs:
	}

	return err
}

func serveHTTP(cfg profile.Server, srv *http.Server, errs chan<- error) {
	srv.Addr = cfg.Addr
	if vhosts := cfg.Vhosts; len(vhosts) != 0 { // 处理 vhosts
		h := srv.Handler
		vhost := ship.NewHostManagerHandler(nil)
		for _, host := range vhosts {
			if _, err := vhost.AddHost(host, h); err != nil {
				errs <- err
				return
			}
		}
		srv.Handler = vhost
	}

	cert, pkey := cfg.Cert, cfg.Pkey
	if cert == "" || pkey == "" {
		errs <- srv.ListenAndServe()
	} else {
		errs <- srv.ListenAndServeTLS(cert, pkey)
	}
}
