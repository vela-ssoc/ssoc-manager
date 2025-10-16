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

	"github.com/vela-ssoc/ssoc-common-mb/cronv3"
	"github.com/vela-ssoc/ssoc-common-mb/dal/gridfs"
	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/gopool"
	"github.com/vela-ssoc/ssoc-common-mb/httpx"
	"github.com/vela-ssoc/ssoc-common-mb/integration/cmdb"
	"github.com/vela-ssoc/ssoc-common-mb/integration/dong/v2"
	"github.com/vela-ssoc/ssoc-common-mb/integration/elastic"
	"github.com/vela-ssoc/ssoc-common-mb/integration/sonatype"
	"github.com/vela-ssoc/ssoc-common-mb/integration/vulnsync"
	"github.com/vela-ssoc/ssoc-common-mb/problem"
	"github.com/vela-ssoc/ssoc-common-mb/profile"
	"github.com/vela-ssoc/ssoc-common-mb/shipx"
	"github.com/vela-ssoc/ssoc-common-mb/sqldb"
	"github.com/vela-ssoc/ssoc-common-mb/storage/v2"
	"github.com/vela-ssoc/ssoc-common-mb/validation"
	"github.com/vela-ssoc/ssoc-manager/app/brkapi"
	"github.com/vela-ssoc/ssoc-manager/app/mgtapi"
	"github.com/vela-ssoc/ssoc-manager/app/middle"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/app/session"
	"github.com/vela-ssoc/ssoc-manager/bridge/blink"
	"github.com/vela-ssoc/ssoc-manager/bridge/linkhub"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/confload"
	"github.com/vela-ssoc/ssoc-manager/integration/casauth"
	"github.com/vela-ssoc/ssoc-manager/integration/cmdb2"
	"github.com/vela-ssoc/ssoc-manager/integration/oauth"
	"github.com/vela-ssoc/vela-common-mba/netutil"
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
	db, err := sqldb.Open(dbCfg.DSN, gormCfg)
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
	valid := validation.New()
	if err = valid.RegisterCustomValidations(validation.Extensions()); err != nil {
		return err
	}

	sh := ship.Default()
	sh.Logger = shipx.NewLog(log)
	sh.Session = sess
	sh.Validator = valid
	sh.NotFound = prob.NotFound
	sh.HandleError = prob.HandleError

	// 静态资源代理
	srvCfg := cfg.Server
	for k, v := range srvCfg.Static {
		if k != "" && v != "" {
			sh.Route(k).Static(v)
		}
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
	pool := gopool.New(4096)
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
	brkHandle := blink.New(huber)           // 将 broker 网关注入到 blink service 中
	blinkREST := mgtapi.NewBlink(brkHandle) // 构造 REST 层
	blinkREST.Route(anon, bearer, basic)    // 注册路由用于调用
	if err = huber.ResetDB(); err != nil {
		return err
	}
	// ==========[ broker end ] ==========

	httpClient := httpx.NewClient()
	casClient := casauth.NewClient(confload.NewCAS(cfg.Oauth.CAS), httpClient, log)

	emcService := service.NewEmc(qry, pusher)
	emcREST := mgtapi.NewEmc(emcService)
	emcREST.Route(anon, bearer, basic)
	store := storage.NewStore(qry)

	digestService := service.NewDigest()
	sequenceService := service.Sequence()

	userService := service.NewUser(qry, digestService, casClient, log)
	userREST := mgtapi.NewUser(userService)
	userREST.Route(anon, bearer, basic)

	if dbCfg.Migrate {
		// 尝试初始化管理员
		if err = userService.Generate(ctx); err != nil {
			return err
		}
	}

	loginLockService := service.NewLoginLock(qry, time.Hour, 10) // 每小时错误 10 次就锁定账户

	oauthCfg := confload.NewOauth(cfg.Oauth.URL, cfg.Oauth.ClientID, cfg.Oauth.ClientSecret, cfg.Oauth.RedirectURL)
	oauthClient := oauth.NewClient(oauthCfg, httpClient, log)
	authService := service.NewAuth(qry, loginLockService, userService, oauthClient)
	authREST := mgtapi.NewAuth(authService)
	authREST.Route(anon, bearer, basic)

	minionFilterSvc, err := service.NewMinionFilter(qry)
	if err != nil {
		return err
	}

	cmdbCfg := cmdb.NewConfigure(store)
	cmdbClient := cmdb.NewClient(qry, cmdbCfg, client)
	minionService := service.NewMinion(qry, cmdbClient, pusher, minionFilterSvc)
	minionREST := mgtapi.NewMinion(qry, huber, minionService)
	minionREST.Route(anon, bearer, basic)

	intoService := service.Into(qry, huber)
	intoREST := mgtapi.NewInto(intoService, headerKey, queryKey)
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
		crond.Schedule("task-timeout-clean", cronv3.NewPeriodicallyTimes(30*time.Second), func() {
			cctx, cancel := context.WithTimeout(context.Background(), time.Hour)
			defer cancel()
			taskExecuteSvc.TimeoutMonitor(cctx)
		})
	}

	taskExtensionSvc := service.NewTaskExtension(qry, huber, minionFilterSvc, crond)
	taskExtensionSvc.Init(ctx) // 启动自动加入任务
	taskExtensionAPI := mgtapi.NewTaskExtension(taskExtensionSvc)
	taskExtensionAPI.Route(anon, bearer, basic)

	tagService := service.NewTag(qry, pusher)
	tagREST := mgtapi.NewTag(tagService)
	tagREST.Route(anon, bearer, basic)

	substanceTaskService := service.NewSubstanceTask(qry, sequenceService, pusher)

	// -----[ 配置与发布 ]-----
	substanceService := service.NewSubstance(qry, pusher, digestService, substanceTaskService)
	substanceREST := mgtapi.NewSubstance(substanceService)
	substanceREST.Route(anon, bearer, basic)

	effectService := service.NewEffect(qry, pusher, sequenceService, substanceTaskService)
	effectREST := mgtapi.NewEffect(effectService)
	effectREST.Route(anon, bearer, basic)

	substanceTaskREST := mgtapi.NewSubstanceTask(qry, substanceTaskService)
	substanceTaskREST.Route(anon, bearer, basic)
	// -----[ 配置与发布 ]-----

	esForwardCfg := elastic.NewConfigure(qry, name)
	esForward := elastic.NewSearch(esForwardCfg, client)
	elasticService := service.NewElastic(qry, pusher, esForward, esForwardCfg, client)
	elasticREST := mgtapi.NewElastic(elasticService, headerKey, queryKey)
	elasticREST.Route(anon, bearer, basic)

	processService := service.NewProcess(qry)
	processREST := mgtapi.NewProcess(processService)
	processREST.Route(anon, bearer, basic)

	alertServerREST := mgtapi.NewAlertServer(alertServerSvc)
	alertServerREST.Route(anon, bearer, basic)
	siemServerREST := mgtapi.NewSIEMServer(siemServerSvc)
	siemServerREST.Route(anon, bearer, basic)

	accountService := service.NewAccount(qry)
	accountREST := mgtapi.NewAccount(accountService)
	accountREST.Route(anon, bearer, basic)

	oplogService := service.NewOplog(qry)
	oplogREST := mgtapi.NewOplog(oplogService)
	oplogREST.Route(anon, bearer, basic)

	notifierService := service.NewNotifier(qry, pusher)
	notifierREST := mgtapi.NewNotifier(notifierService)
	notifierREST.Route(anon, bearer, basic)

	minionTaskService := service.NewMinionTask(qry, log)
	minionTaskREST := mgtapi.NewMinionTask(minionTaskService)
	minionTaskREST.Route(anon, bearer, basic)

	es := elastic.NewSearch(elastic.NewConfigure(qry, "ES"), client)
	minionLogonService := service.NewMinionLogon(qry, es)
	minionLogonREST := mgtapi.NewMinionLogon(minionLogonService)
	minionLogonREST.Route(anon, bearer, basic)

	riskService := service.NewRisk(qry, store)
	riskREST := mgtapi.NewRisk(riskService)
	riskREST.Route(anon, bearer, basic)

	passDNSService := service.NewPassDNS(qry)
	passDNSREST := mgtapi.NewPassDNS(passDNSService)
	passDNSREST.Route(anon, bearer, basic)
	passIPService := service.NewPassIP(qry)
	passIPREST := mgtapi.NewPassIP(passIPService)
	passIPREST.Route(anon, bearer, basic)
	riskDNSService := service.NewRiskDNS(qry)
	riskDNSREST := mgtapi.NewRiskDNS(riskDNSService)
	riskDNSREST.Route(anon, bearer, basic)
	riskFileService := service.NewRiskFile(qry)
	riskFileREST := mgtapi.NewRiskFile(riskFileService)
	riskFileREST.Route(anon, bearer, basic)

	storeService := service.NewStore(qry, pusher, store)
	eventService := service.NewEvent(qry, store)
	eventREST := mgtapi.NewEvent(eventService)
	eventREST.Route(anon, bearer, basic)
	storeREST := mgtapi.NewStore(storeService)
	storeREST.Route(anon, bearer, basic)

	sbomComponentService := service.NewSBOMComponent(qry)
	sbomComponentREST := mgtapi.NewSBOMComponent(sbomComponentService)
	sbomComponentREST.Route(anon, bearer, basic)
	sbomProjectService := service.NewSBOMProject(qry)
	sbomProjectREST := mgtapi.NewSBOMProject(sbomProjectService)
	sbomProjectREST.Route(anon, bearer, basic)
	sbomVulnService := service.NewSBOMVuln(qry)
	sbomVulnREST := mgtapi.NewSBOMVuln(sbomVulnService)
	sbomVulnREST.Route(anon, bearer, basic)

	vipService := service.NewVIP(qry)
	vipREST := mgtapi.NewVIP(vipService)
	vipREST.Route(anon, bearer, basic)

	cmdbService := service.NewCmdb(qry)
	cmdbREST := mgtapi.NewCmdb(cmdbService)
	cmdbREST.Route(anon, bearer, basic)

	dashService := service.NewDash(qry)
	dashREST := mgtapi.NewDash(dashService)
	dashREST.Route(anon, bearer, basic)

	thirdService := service.NewThird(qry, pusher, gfs)
	thirdREST := mgtapi.NewThird(thirdService)
	thirdREST.Route(anon, bearer, basic)

	thirdCustomizedService := service.NewThirdCustomized(qry)
	thirdCustomizedREST := mgtapi.NewThirdCustomized(thirdCustomizedService)
	thirdCustomizedREST.Route(anon, bearer, basic)

	brokerService := service.NewBroker(qry)
	brokerREST := mgtapi.Broker(brokerService)
	brokerREST.Route(anon, bearer, basic)

	brokerBinaryService := service.NewBrokerBinary(qry, gfs, store)
	brokerBinaryREST := mgtapi.NewBrokerBinary(brokerBinaryService)
	brokerBinaryREST.Route(anon, bearer, basic)

	brokerSystemSvc := service.NewBrokerSystem(huber)
	brokerSystemAPI := mgtapi.NewBrokerSystem(brokerSystemSvc)
	brokerSystemAPI.Route(anon, bearer, basic)

	certService := service.NewCert(qry)
	certREST := mgtapi.NewCert(certService)
	certREST.Route(anon, bearer, basic)

	minionBinaryService := service.NewMinionBinary(qry, pusher, gfs)
	minionBinaryREST := mgtapi.NewMinionBinary(minionBinaryService)
	minionBinaryREST.Route(anon, bearer, basic)

	minionListenService := service.NewMinionListen(qry)
	minionListenREST := mgtapi.NewMinionListen(minionListenService)
	minionListenREST.Route(anon, bearer, basic)

	minionAccountService := service.NewMinionAccount(qry)
	minionAccountREST := mgtapi.NewMinionAccount(minionAccountService)
	minionAccountREST.Route(anon, bearer, basic)

	deployService := service.NewDeploy(qry, store, gfs)
	deployREST := mgtapi.NewDeploy(deployService)
	deployREST.Route(anon, bearer, basic)

	domainService := service.NewDomain(qry)
	domainREST := mgtapi.NewDomain(domainService)
	domainREST.Route(anon, bearer, basic)

	riskIPService := service.NewRiskIP(qry)
	riskIPREST := mgtapi.NewRiskIP(riskIPService)
	riskIPREST.Route(anon, bearer, basic)

	minionCustomizedService := service.NewMinionCustomized(qry)
	minionCustomizedREST := mgtapi.NewMinionCustomized(minionCustomizedService)
	minionCustomizedREST.Route(anon, bearer, basic)

	emailService := service.NewEmail(qry, pusher)
	emailREST := mgtapi.NewEmail(emailService)
	emailREST.Route(anon, bearer, basic)

	startupService := service.NewStartup(qry, store, pusher)
	startupREST := mgtapi.NewStartup(startupService)
	startupREST.Route(anon, bearer, basic)

	sharedService := service.NewShared(qry)
	sharedAPI := mgtapi.NewShared(sharedService)
	sharedAPI.Route(anon, bearer, basic)

	hardConfig := sonatype.HardConfig()
	sona := sonatype.NewClient(hardConfig, client)
	synchro := vulnsync.New(db, sona)
	mgtapi.NewManual(synchro).Route(anon, bearer, basic)

	cmdb2Config := confload.NewCmdb2(cfg.Cmdb2.URL, cfg.Cmdb2.AccessKey, cfg.Cmdb2.SecretKey)
	cmdb2Client := cmdb2.NewClient(cmdb2Config, httpClient)
	cmdb2Service := service.NewCmdb2(qry, cmdb2Client)
	mgtapi.NewCmdb2(cmdb2Service).Route(anon, bearer, basic)

	davREST := mgtapi.NewDavFS(base)
	davREST.Route(anon, bearer, basic)

	pprofDir := "resources/pprof"
	_ = os.RemoveAll(pprofDir)
	_ = os.MkdirAll(pprofDir, os.ModePerm)
	pprofService := service.NewPprof(qry, pprofDir, pusher)
	pprofREST := mgtapi.NewPprof(pprofService)
	pprofREST.Route(anon, bearer, basic)

	app := &application{
		cfg:     cfg,
		handler: sh,
		parent:  ctx,
	}

	return app.run()
}
