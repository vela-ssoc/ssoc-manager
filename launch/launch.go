package launch

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vela-ssoc/vela-common-mb/cronv3"
	"github.com/vela-ssoc/vela-common-mb/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/gopool"
	"github.com/vela-ssoc/vela-common-mb/httpx"
	"github.com/vela-ssoc/vela-common-mb/integration/cmdb"
	"github.com/vela-ssoc/vela-common-mb/integration/dong/v2"
	"github.com/vela-ssoc/vela-common-mb/integration/elastic"
	"github.com/vela-ssoc/vela-common-mb/integration/sonatype"
	"github.com/vela-ssoc/vela-common-mb/integration/vulnsync"
	"github.com/vela-ssoc/vela-common-mb/problem"
	"github.com/vela-ssoc/vela-common-mb/profile"
	"github.com/vela-ssoc/vela-common-mb/shipx"
	"github.com/vela-ssoc/vela-common-mb/sqldb"
	"github.com/vela-ssoc/vela-common-mb/storage/v2"
	"github.com/vela-ssoc/vela-common-mb/validate"
	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-manager/app/brkapi"
	"github.com/vela-ssoc/vela-manager/app/mgtapi"
	"github.com/vela-ssoc/vela-manager/app/middle"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/app/session"
	"github.com/vela-ssoc/vela-manager/bridge/blink"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/confload"
	"github.com/vela-ssoc/vela-manager/integration/casauth"
	"github.com/vela-ssoc/vela-manager/integration/cmdb2"
	"github.com/vela-ssoc/vela-manager/integration/oauth"
	"github.com/xgfone/ship/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Run(ctx context.Context, path string) error {
	cfg, err := profile.ParseManager(path)
	if err != nil {
		return err
	}

	return runApp(ctx, cfg)
}

func runApp(ctx context.Context, cfg *profile.ManagerConfig) error {
	logCfg := cfg.Logger
	//goland:noinspection GoUnhandledErrorResult
	defer logCfg.Close()

	logWriter := logCfg.LogWriter()
	logOption := &slog.HandlerOptions{AddSource: true, Level: logWriter.Level()}
	logHandler := slog.NewJSONHandler(logWriter, logOption)
	log := slog.New(logHandler)
	log.Info("日志组件初始化完毕")

	dbCfg := cfg.Database
	gormLogLevel := sqldb.MappingGormLogLevel(dbCfg.Level)
	gormLog, _ := sqldb.NewLog(logWriter, logger.Config{LogLevel: gormLogLevel})
	gormCfg := &gorm.Config{Logger: gormLog}
	db, err := sqldb.Open(dbCfg.DSN, log, gormCfg)
	if err != nil {
		return err
	}
	sdb, err := db.DB()
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer sdb.Close() // 程序结束时断开数据库连接。

	sdb.SetMaxOpenConns(dbCfg.MaxOpenConn)
	sdb.SetMaxIdleConns(dbCfg.MaxIdleConn)
	sdb.SetConnMaxLifetime(dbCfg.MaxLifeTime.Duration())
	sdb.SetConnMaxIdleTime(dbCfg.MaxIdleTime.Duration())
	log.Warn("当前数据库类型", slog.String("dialect", db.Dialector.Name()))

	if dbCfg.Migrate { // 迁移合并数据库
		log.Info("开始执行数据库迁移合并")
		if err = db.WithContext(ctx).AutoMigrate(model.All()...); err != nil {
			return err
		}
		log.Info("数据库迁移合并执行完成")
	}

	qry := query.Use(db)
	var gfs gridfs.FS
	if dir := cfg.Server.CDN; dir == "" {
		gfs = gridfs.NewFS(qry)
	} else {
		cdn := filepath.Clean(dir)
		if err = os.MkdirAll(cdn, os.ModePerm); err != nil {
			return err
		}
		gfs = gridfs.NewCache(qry, cdn)
	}

	const name = "manager"
	const headerKey = ship.HeaderAuthorization
	queryKey := strings.ToLower(headerKey)
	auth := middle.Auth(headerKey, queryKey)
	routeRecord := route.NewRecord(qry)
	recordMid := middle.Oplog(routeRecord)

	prob := problem.NewHandle(name)
	sess := session.DBSess(qry, cfg.Server.Session.Duration())
	valid := validate.New()

	sh := ship.Default()
	sh.Logger = shipx.NewLog(log)
	sh.Session = sess
	sh.Validator = valid
	sh.NotFound = prob.NotFound
	sh.HandleError = prob.HandleError

	// 静态资源代理
	srvCfg := cfg.Server
	if static := srvCfg.Static; static != "" {
		sh.Route("/").Static(static)
	}

	base := "/api/v1"
	anon := sh.Group(base).Use(recordMid)
	bearer := anon.Clone().Use(auth.Bearer)
	basic := anon.Clone().Use(auth.Basic)

	alertServerSvc := service.NewAlertServer(qry)
	siemServerSvc := service.NewSIEMServer(qry)

	client := netutil.NewClient()
	dongCli := dong.NewAlert(alertServerSvc)

	// 初始化协程池
	pool := gopool.NewV2(8192)
	crond := cronv3.New(log)
	crond.Start()
	defer crond.Stop()
	{
		buf := make([]byte, 50)
		_, _ = rand.Read(buf)
		bgName := "background-cleanup-" + hex.EncodeToString(buf)
		spec := cronv3.NewPeriodicallyTimes(time.Hour)
		crond.Schedule(bgName, spec, crond.Cleanup)
	}

	// ==========[ broker begin ] ==========
	brkmux := ship.Default()
	brkmux.Validator = valid
	brkmux.NotFound = prob.NotFound
	brkmux.HandleError = prob.HandleError
	brkgrp := brkmux.Group("/")
	{
		alert := brkapi.NewAlert(dongCli, log)
		alert.Router(brkgrp)
		siemAPI := brkapi.NewSIEM(siemServerSvc)
		siemAPI.Router(brkgrp)
	}

	huber := linkhub.New(qry, brkmux, pool, cfg) // 将连接中心注入到 broker 接入网关中
	pusher := push.NewPush(qry, huber)
	brkHandle := blink.New(huber)        // 将 broker 网关注入到 blink service 中
	blinkREST := mgtapi.Blink(brkHandle) // 构造 REST 层
	blinkREST.Route(anon, bearer, basic) // 注册路由用于调用
	if err = huber.ResetDB(); err != nil {
		return err
	}
	// ==========[ broker end ] ==========

	httpClient := httpx.NewClient()
	casClient := casauth.NewClient(confload.NewCAS(cfg.Oauth.CAS), httpClient, log)

	emcService := service.Emc(qry, pusher)
	emcREST := mgtapi.Emc(emcService)
	emcREST.Route(anon, bearer, basic)
	store := storage.NewStore(qry)

	digestService := service.Digest()
	sequenceService := service.Sequence()

	userService := service.User(qry, digestService, casClient, log)
	userREST := mgtapi.User(userService)
	userREST.Route(anon, bearer, basic)

	if dbCfg.Migrate {
		// 尝试初始化管理员
		if err = userService.Generate(ctx); err != nil {
			return err
		}
	}

	verifyService := service.Verify(qry, 3, store, log)       // 验证码 3 分钟有效期
	loginLockService := service.LoginLock(qry, time.Hour, 10) // 每小时错误 10 次就锁定账户

	oauthCfg := confload.NewOauth(cfg.Oauth.URL, cfg.Oauth.ClientID, cfg.Oauth.ClientSecret, cfg.Oauth.RedirectURL)
	oauthClient := oauth.NewClient(oauthCfg, httpClient, log)
	authService := service.Auth(qry, verifyService, loginLockService, userService, oauthClient)
	authREST := mgtapi.Auth(authService)
	authREST.Route(anon, bearer, basic)

	minionFilterSvc, err := service.NewMinionFilter(qry)
	if err != nil {
		return err
	}

	cmdbCfg := cmdb.NewConfigure(store)
	cmdbClient := cmdb.NewClient(qry, cmdbCfg, client)
	minionService := service.Minion(qry, cmdbClient, pusher, minionFilterSvc)
	minionREST := mgtapi.Minion(qry, huber, minionService)
	minionREST.Route(anon, bearer, basic)

	intoService := service.Into(qry, huber)
	intoREST := mgtapi.Into(intoService, headerKey, queryKey)
	intoREST.Route(anon, bearer, basic)

	luaTemplateSvc := service.NewLuaTemplate()
	luaTemplateAPI := mgtapi.NewLuaTemplate(luaTemplateSvc)
	luaTemplateAPI.Route(anon, bearer, basic)

	extensionMarketSvc := service.NewExtensionMarket(qry)
	extensionMarketAPI := mgtapi.NewExtensionMarket(extensionMarketSvc)
	extensionMarketAPI.Route(anon, bearer, basic)

	taskExecuteSvc := service.NewTaskExecute(qry, log)
	taskExecuteAPI := mgtapi.NewTaskExecute(taskExecuteSvc)
	taskExecuteAPI.Route(anon, bearer, basic)

	taskExecuteItemSvc, err := service.NewTaskExecuteItem(qry)
	if err != nil {
		return err
	}
	taskExecuteItemAPI := mgtapi.NewTaskExecuteItem(taskExecuteItemSvc)
	taskExecuteItemAPI.Route(anon, bearer, basic)

	{
		crond.Schedule("task-timeout-clean", cronv3.NewPeriodicallyTimes(5*time.Minute), func() {
			cctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()
			taskExecuteSvc.TimeoutMonitor(cctx)
		})
	}

	taskExtensionSvc := service.NewTaskExtension(qry, huber, minionFilterSvc, crond)
	taskExtensionAPI := mgtapi.NewTaskExtension(taskExtensionSvc)
	taskExtensionAPI.Route(anon, bearer, basic)

	tagService := service.Tag(qry, pusher)
	tagREST := mgtapi.Tag(tagService)
	tagREST.Route(anon, bearer, basic)

	substanceTaskService := service.SubstanceTask(qry, sequenceService, pusher)

	// -----[ 配置与发布 ]-----
	substanceService := service.Substance(qry, pusher, digestService, substanceTaskService)
	substanceREST := mgtapi.Substance(substanceService)
	substanceREST.Route(anon, bearer, basic)

	effectService := service.Effect(qry, pusher, sequenceService, substanceTaskService)
	effectREST := mgtapi.Effect(effectService)
	effectREST.Route(anon, bearer, basic)

	substanceTaskREST := mgtapi.SubstanceTask(qry, substanceTaskService)
	substanceTaskREST.Route(anon, bearer, basic)
	// -----[ 配置与发布 ]-----

	esForwardCfg := elastic.NewConfigure(qry, name)
	esForward := elastic.NewSearch(esForwardCfg, client)
	elasticService := service.Elastic(qry, pusher, esForward, esForwardCfg, client)
	elasticREST := mgtapi.Elastic(elasticService, headerKey, queryKey)
	elasticREST.Route(anon, bearer, basic)

	processService := service.Process(qry)
	processREST := mgtapi.Process(processService)
	processREST.Route(anon, bearer, basic)

	alertServerREST := mgtapi.NewAlertServer(alertServerSvc)
	alertServerREST.Route(anon, bearer, basic)
	siemServerREST := mgtapi.NewSIEMServer(siemServerSvc)
	siemServerREST.Route(anon, bearer, basic)

	accountService := service.Account(qry)
	accountREST := mgtapi.Account(accountService)
	accountREST.Route(anon, bearer, basic)

	oplogService := service.Oplog(qry)
	oplogREST := mgtapi.Oplog(oplogService)
	oplogREST.Route(anon, bearer, basic)

	notifierService := service.Notifier(qry, pusher)
	notifierREST := mgtapi.Notifier(notifierService)
	notifierREST.Route(anon, bearer, basic)

	minionTaskService := service.MinionTask(qry)
	minionTaskREST := mgtapi.MinionTask(minionTaskService)
	minionTaskREST.Route(anon, bearer, basic)

	es := elastic.NewSearch(elastic.NewConfigure(qry, "ES"), client)
	minionLogonService := service.MinionLogon(qry, es)
	minionLogonREST := mgtapi.MinionLogon(minionLogonService)
	minionLogonREST.Route(anon, bearer, basic)

	riskService := service.Risk(qry, store)
	riskREST := mgtapi.Risk(qry, riskService)
	riskREST.Route(anon, bearer, basic)

	passDNSService := service.PassDNS(qry)
	passDNSREST := mgtapi.PassDNS(passDNSService)
	passDNSREST.Route(anon, bearer, basic)
	passIPService := service.PassIP(qry)
	passIPREST := mgtapi.PassIP(passIPService)
	passIPREST.Route(anon, bearer, basic)
	riskDNSService := service.RiskDNS(qry)
	riskDNSREST := mgtapi.RiskDNS(riskDNSService)
	riskDNSREST.Route(anon, bearer, basic)
	riskFileService := service.RiskFile(qry)
	riskFileREST := mgtapi.RiskFile(riskFileService)
	riskFileREST.Route(anon, bearer, basic)

	storeService := service.Store(qry, pusher, store)
	eventService := service.Event(qry, store)
	eventREST := mgtapi.Event(eventService)
	eventREST.Route(anon, bearer, basic)
	storeREST := mgtapi.Store(storeService)
	storeREST.Route(anon, bearer, basic)

	sbomComponentService := service.SBOMComponent(qry)
	sbomComponentREST := mgtapi.SBOMComponent(sbomComponentService)
	sbomComponentREST.Route(anon, bearer, basic)
	sbomProjectService := service.SBOMProject(qry)
	sbomProjectREST := mgtapi.SBOMProject(sbomProjectService)
	sbomProjectREST.Route(anon, bearer, basic)
	sbomVulnService := service.SBOMVuln(qry)
	sbomVulnREST := mgtapi.SBOMVuln(sbomVulnService)
	sbomVulnREST.Route(anon, bearer, basic)

	vipService := service.VIP(qry)
	vipREST := mgtapi.VIP(vipService)
	vipREST.Route(anon, bearer, basic)

	cmdbService := service.Cmdb(qry)
	cmdbREST := mgtapi.Cmdb(cmdbService)
	cmdbREST.Route(anon, bearer, basic)

	dashService := service.Dash(qry)
	dashREST := mgtapi.Dash(dashService)
	dashREST.Route(anon, bearer, basic)

	thirdService := service.Third(qry, pusher, gfs)
	thirdREST := mgtapi.Third(thirdService)
	thirdREST.Route(anon, bearer, basic)

	thirdCustomizedService := service.ThirdCustomized(qry)
	thirdCustomizedREST := mgtapi.ThirdCustomized(thirdCustomizedService)
	thirdCustomizedREST.Route(anon, bearer, basic)

	brokerService := service.Broker(qry)
	brokerREST := mgtapi.Broker(brokerService)
	brokerREST.Route(anon, bearer, basic)

	brokerBinaryService := service.BrokerBinary(qry, gfs, store)
	brokerBinaryREST := mgtapi.BrokerBinary(brokerBinaryService)
	brokerBinaryREST.Route(anon, bearer, basic)

	brokerCommandSvc := service.NewBrokerCommand(huber)
	brokerCommandAPI := mgtapi.NewBrokerCommand(brokerCommandSvc)
	brokerCommandAPI.Route(anon, bearer, basic)

	certService := service.Cert(qry)
	certREST := mgtapi.Cert(certService)
	certREST.Route(anon, bearer, basic)

	minionBinaryService := service.MinionBinary(qry, pusher, gfs)
	minionBinaryREST := mgtapi.MinionBinary(minionBinaryService)
	minionBinaryREST.Route(anon, bearer, basic)

	minionListenService := service.MinionListen(qry)
	minionListenREST := mgtapi.MinionListen(minionListenService)
	minionListenREST.Route(anon, bearer, basic)

	minionAccountService := service.MinionAccount(qry)
	minionAccountREST := mgtapi.MinionAccount(minionAccountService)
	minionAccountREST.Route(anon, bearer, basic)

	deployService := service.Deploy(qry, store, gfs)
	deployREST := mgtapi.Deploy(deployService)
	deployREST.Route(anon, bearer, basic)

	domainService := service.Domain(qry)
	domainREST := mgtapi.Domain(domainService)
	domainREST.Route(anon, bearer, basic)

	riskIPService := service.RiskIP(qry)
	riskIPREST := mgtapi.RiskIP(riskIPService)
	riskIPREST.Route(anon, bearer, basic)

	minionCustomizedService := service.MinionCustomized(qry)
	minionCustomizedREST := mgtapi.MinionCustomized(minionCustomizedService)
	minionCustomizedREST.Route(anon, bearer, basic)

	// mgtapi.NewMinionFilter(minionFilterSvc).Route(anon, bearer, basic)

	emailService := service.Email(qry, pusher)
	emailREST := mgtapi.Email(emailService)
	emailREST.Route(anon, bearer, basic)

	startupService := service.Startup(qry, store, pusher)
	startupREST := mgtapi.Startup(startupService)
	startupREST.Route(anon, bearer, basic)

	sharedService := service.Shared(qry)
	sharedAPI := mgtapi.Shared(sharedService)
	sharedAPI.Route(anon, bearer, basic)

	hardConfig := sonatype.HardConfig()
	sona := sonatype.NewClient(hardConfig, client)
	synchro := vulnsync.New(db, sona)
	mgtapi.Manual(synchro).Route(anon, bearer, basic)

	cmdb2Config := confload.NewCmdb2(cfg.Cmdb2.URL, cfg.Cmdb2.AccessKey, cfg.Cmdb2.SecretKey)
	cmdb2Client := cmdb2.NewClient(cmdb2Config, httpClient)
	cmdb2Service := service.Cmdb2(qry, cmdb2Client)
	mgtapi.Cmdb2(cmdb2Service).Route(anon, bearer, basic)

	davREST := mgtapi.DavFS(base)
	davREST.Route(anon, bearer, basic)

	pprofDir := "resources/pprof"
	_ = os.RemoveAll(pprofDir)
	_ = os.MkdirAll(pprofDir, os.ModePerm)
	pprofService := service.Pprof(qry, pprofDir, pusher)
	pprofREST := mgtapi.Pprof(pprofService)
	pprofREST.Route(anon, bearer, basic)

	app := &application{
		cfg:     cfg,
		handler: sh,
		parent:  ctx,
	}

	return app.run()
}
