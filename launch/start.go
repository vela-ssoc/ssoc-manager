package launch

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/dal/entity"
	"github.com/vela-ssoc/vela-common-mb/dal/gridfs2"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/sqldb"
	"github.com/vela-ssoc/vela-common-mb/validate"
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
		log.Info("当前连接的是 OpenGauss 信创数据库")
	} else {
		log.Info("当前连接的是 MySQL 数据库")
	}
	if dbCfg.Migrate { // 迁移合并数据库
		if err = entity.Migrate(db); err != nil {
			return err
		}
	}

	qry := query.Use(db)
	gridFS := gridfs.New(qry)

	alertServerSvc := service.NewAlertServer(qry, log)
	gridSvc := service.NewGrid(gridFS, qry, log)
	logSvc := service.NewLog(logLevel, gormLevel, log)
	mgtRouters := []shipx.Router{
		restapi.NewAlertServer(alertServerSvc),
		restapi.NewGrid(gridSvc),
		restapi.NewLog(logSvc),
		restapi.NewSystem(),
	}
	mgtMux := ship.Default()
	mgtMux.Validator = valid
	mgtMux.Logger = shipx.NewLog(log)
	mgtMux.HandleError = shipx.HandleError
	mgtRGB := mgtMux.Group("/api/v1")
	if err = shipx.BindRouters(mgtRGB, mgtRouters); err != nil {
		return err
	}

	srvCfg := cfg.Server
	mgtSrv := &http.Server{Handler: mgtMux}
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
