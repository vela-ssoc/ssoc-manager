package launch

import (
	"context"
	stdlog "log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/vela-ssoc/vela-common-mb/cmdb2"
	"github.com/vela-ssoc/vela-common-mb/dal/gridfs"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/gopool"
	"github.com/vela-ssoc/vela-common-mb/integration/cmdb"
	"github.com/vela-ssoc/vela-common-mb/integration/dong/v2"
	"github.com/vela-ssoc/vela-common-mb/integration/elastic"
	"github.com/vela-ssoc/vela-common-mb/integration/sonatype"
	"github.com/vela-ssoc/vela-common-mb/integration/ssoauth"
	"github.com/vela-ssoc/vela-common-mb/integration/vulnsync"
	"github.com/vela-ssoc/vela-common-mb/logback"
	"github.com/vela-ssoc/vela-common-mb/problem"
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
	"github.com/vela-ssoc/vela-manager/httpx"
	"github.com/vela-ssoc/vela-manager/infra/config"
	"github.com/vela-ssoc/vela-manager/infra/profile"
	"github.com/vela-ssoc/vela-manager/oauth2"
	"github.com/xgfone/ship/v5"
)

func Run(ctx context.Context, path string, slog logback.Logger) error {
	var cfg config.Config
	if err := profile.Load(path, &cfg); err != nil {
		return err
	}

	app, err := newApp(ctx, cfg, slog)
	if err != nil {
		return err
	}

	return app.run()
}

func newApp(ctx context.Context, cfg config.Config, slog logback.Logger) (*application, error) {
	dbCfg := cfg.Database
	// logCfg := cfg.Logger

	zapLog := cfg.Logger.Zap()
	slog.Replace(zapLog)
	// gormLog := logback.Gorm(zapLog, logCfg.Level)

	ormLog, _ := sqldb.NewLog(os.Stdout, logger.Config{LogLevel: logger.Info})
	ormCfg := &gorm.Config{Logger: ormLog}
	db, _, err := sqldb.Open(dbCfg.DSN, stdlog.Default(), ormCfg)
	if err != nil {
		return nil, err
	}
	sdb, err := db.DB()
	if err != nil {
		return nil, err
	}

	//db, sdb, err := dbms.Open(dbCfg, gormLog)
	//if err != nil {
	//	return nil, err
	//}
	tables := []any{
		model.AlertServer{},
		model.SIEMServer{},
	}
	if err = db.AutoMigrate(tables...); err != nil {
		return nil, err
	}

	query.SetDefault(db)
	qry := query.Q
	secCfg := cfg.Section

	var gfs gridfs.FS
	if dir := secCfg.CDN; dir == "" {
		gfs = gridfs.NewFS(sdb)
	} else {
		cdn := filepath.Clean(dir)
		if err = os.MkdirAll(cdn, os.ModePerm); err != nil {
			return nil, err
		}
		gfs = gridfs.NewCache(sdb, cdn)
	}

	const name = "manager"
	const headerKey = ship.HeaderAuthorization
	queryKey := strings.ToLower(headerKey)
	auth := middle.Auth(headerKey, queryKey)
	routeRecord := route.NewRecord()
	recordMid := middle.Oplog(routeRecord)

	prob := problem.NewHandle(name)
	sess := session.DBSess(secCfg.Sess)
	valid := validate.New()

	sh := ship.Default()
	sh.Logger = slog
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

	// ==========[ broker begin ] ==========
	brkmux := ship.Default()
	brkmux.Validator = valid
	brkmux.NotFound = prob.NotFound
	brkmux.HandleError = prob.HandleError
	brkgrp := brkmux.Group("/")
	{
		alert := brkapi.NewAlert(dongCli, slog)
		alert.Router(brkgrp)
		siemAPI := brkapi.NewSIEM(siemServerSvc)
		siemAPI.Router(brkgrp)
	}

	huber := linkhub.New(brkmux, pool, cfg) // 将连接中心注入到 broker 接入网关中
	pusher := push.NewPush(huber)
	brkHandle := blink.New(huber)        // 将 broker 网关注入到 blink service 中
	blinkREST := mgtapi.Blink(brkHandle) // 构造 REST 层
	blinkREST.Route(anon, bearer, basic) // 注册路由用于调用
	if err = huber.ResetDB(); err != nil {
		return nil, err
	}
	// ==========[ broker end ] ==========

	emcService := service.Emc(pusher)
	emcREST := mgtapi.Emc(emcService)
	emcREST.Route(anon, bearer, basic)
	store := storage.NewStore()

	digestService := service.Digest()
	sequenceService := service.Sequence()

	ssoCfg := ssoauth.NewConfigure(store)
	ssoCli := ssoauth.NewClient(ssoCfg, client, slog)
	userService := service.User(digestService, ssoCli)
	userREST := mgtapi.User(userService)
	userREST.Route(anon, bearer, basic)

	verifyService := service.Verify(3, dongCli, store, slog) // 验证码 3 分钟有效期
	loginLockService := service.LoginLock(time.Hour, 10)     // 每小时错误 10 次就锁定账户

	httpxCli := httpx.Client{Client: http.DefaultClient}
	oauthCli := oauth2.NewClient(cfg.Oauth, httpxCli)

	authService := service.Auth(verifyService, loginLockService, userService, oauthCli)
	authREST := mgtapi.Auth(authService)
	authREST.Route(anon, bearer, basic)

	cmdbCfg := cmdb.NewConfigure(store)
	cmdbClient := cmdb.NewClient(cmdbCfg, client, slog)
	minionService := service.Minion(cmdbClient, pusher)
	minionREST := mgtapi.Minion(huber, minionService)
	minionREST.Route(anon, bearer, basic)

	intoService := service.Into(huber)
	intoREST := mgtapi.Into(intoService, headerKey, queryKey)
	intoREST.Route(anon, bearer, basic)

	tagService := service.Tag(pusher)
	tagREST := mgtapi.Tag(tagService)
	tagREST.Route(anon, bearer, basic)

	substanceTaskService := service.SubstanceTask(sequenceService, pusher)

	// -----[ 配置与发布 ]-----
	substanceService := service.Substance(pusher, digestService, substanceTaskService)
	substanceREST := mgtapi.Substance(substanceService)
	substanceREST.Route(anon, bearer, basic)

	effectService := service.Effect(pusher, sequenceService, substanceTaskService)
	effectREST := mgtapi.Effect(effectService)
	effectREST.Route(anon, bearer, basic)

	substanceTaskREST := mgtapi.SubstanceTask(substanceTaskService)
	substanceTaskREST.Route(anon, bearer, basic)
	// -----[ 配置与发布 ]-----

	esForwardCfg := elastic.NewConfigure(name)
	esForward := elastic.NewSearch(esForwardCfg, client)
	elasticService := service.Elastic(pusher, esForward, esForwardCfg, client)
	elasticREST := mgtapi.Elastic(elasticService, headerKey, queryKey)
	elasticREST.Route(anon, bearer, basic)

	processService := service.Process()
	processREST := mgtapi.Process(processService)
	processREST.Route(anon, bearer, basic)

	alertServerREST := mgtapi.NewAlertServer(alertServerSvc)
	alertServerREST.Route(anon, bearer, basic)
	siemServerREST := mgtapi.NewSIEMServer(siemServerSvc)
	siemServerREST.Route(anon, bearer, basic)

	accountService := service.Account()
	accountREST := mgtapi.Account(accountService)
	accountREST.Route(anon, bearer, basic)

	oplogService := service.Oplog()
	oplogREST := mgtapi.Oplog(oplogService)
	oplogREST.Route(anon, bearer, basic)

	notifierService := service.Notifier(pusher)
	notifierREST := mgtapi.Notifier(notifierService)
	notifierREST.Route(anon, bearer, basic)

	minionTaskService := service.MinionTask()
	minionTaskREST := mgtapi.MinionTask(minionTaskService)
	minionTaskREST.Route(anon, bearer, basic)

	es := elastic.NewSearch(elastic.NewConfigure("ES"), client)
	minionLogonService := service.MinionLogon(es)
	minionLogonREST := mgtapi.MinionLogon(minionLogonService)
	minionLogonREST.Route(anon, bearer, basic)

	riskService := service.Risk(store)
	riskREST := mgtapi.Risk(riskService)
	riskREST.Route(anon, bearer, basic)

	passDNSService := service.PassDNS()
	passDNSREST := mgtapi.PassDNS(passDNSService)
	passDNSREST.Route(anon, bearer, basic)
	passIPService := service.PassIP()
	passIPREST := mgtapi.PassIP(passIPService)
	passIPREST.Route(anon, bearer, basic)
	riskDNSService := service.RiskDNS()
	riskDNSREST := mgtapi.RiskDNS(riskDNSService)
	riskDNSREST.Route(anon, bearer, basic)
	riskFileService := service.RiskFile()
	riskFileREST := mgtapi.RiskFile(riskFileService)
	riskFileREST.Route(anon, bearer, basic)

	storeService := service.Store(pusher, store)
	eventService := service.Event(store)
	eventREST := mgtapi.Event(eventService)
	eventREST.Route(anon, bearer, basic)
	storeREST := mgtapi.Store(storeService)
	storeREST.Route(anon, bearer, basic)

	sbomComponentService := service.SBOMComponent()
	sbomComponentREST := mgtapi.SBOMComponent(sbomComponentService)
	sbomComponentREST.Route(anon, bearer, basic)
	sbomProjectService := service.SBOMProject()
	sbomProjectREST := mgtapi.SBOMProject(sbomProjectService)
	sbomProjectREST.Route(anon, bearer, basic)
	sbomVulnService := service.SBOMVuln()
	sbomVulnREST := mgtapi.SBOMVuln(sbomVulnService)
	sbomVulnREST.Route(anon, bearer, basic)

	vipService := service.VIP()
	vipREST := mgtapi.VIP(vipService)
	vipREST.Route(anon, bearer, basic)

	cmdbService := service.Cmdb()
	cmdbREST := mgtapi.Cmdb(cmdbService)
	cmdbREST.Route(anon, bearer, basic)

	dashService := service.Dash()
	dashREST := mgtapi.Dash(dashService)
	dashREST.Route(anon, bearer, basic)

	thirdService := service.Third(pusher, gfs)
	thirdREST := mgtapi.Third(thirdService)
	thirdREST.Route(anon, bearer, basic)

	thirdCustomizedService := service.ThirdCustomized()
	thirdCustomizedREST := mgtapi.ThirdCustomized(thirdCustomizedService)
	thirdCustomizedREST.Route(anon, bearer, basic)

	brokerService := service.Broker()
	brokerREST := mgtapi.Broker(brokerService)
	brokerREST.Route(anon, bearer, basic)

	brokerBinaryService := service.BrokerBinary(gfs, store)
	brokerBinaryREST := mgtapi.BrokerBinary(brokerBinaryService)
	brokerBinaryREST.Route(anon, bearer, basic)

	certService := service.Cert()
	certREST := mgtapi.Cert(certService)
	certREST.Route(anon, bearer, basic)

	minionBinaryService := service.MinionBinary(pusher, gfs)
	minionBinaryREST := mgtapi.MinionBinary(minionBinaryService)
	minionBinaryREST.Route(anon, bearer, basic)

	minionListenService := service.MinionListen()
	minionListenREST := mgtapi.MinionListen(minionListenService)
	minionListenREST.Route(anon, bearer, basic)

	minionAccountService := service.MinionAccount()
	minionAccountREST := mgtapi.MinionAccount(minionAccountService)
	minionAccountREST.Route(anon, bearer, basic)

	deployService := service.Deploy(store, gfs)
	deployREST := mgtapi.Deploy(deployService)
	deployREST.Route(anon, bearer, basic)

	domainService := service.Domain()
	domainREST := mgtapi.Domain(domainService)
	domainREST.Route(anon, bearer, basic)

	riskIPService := service.RiskIP()
	riskIPREST := mgtapi.RiskIP(riskIPService)
	riskIPREST.Route(anon, bearer, basic)

	minionCustomizedService := service.MinionCustomized()
	minionCustomizedREST := mgtapi.MinionCustomized(minionCustomizedService)
	minionCustomizedREST.Route(anon, bearer, basic)

	emailService := service.Email(pusher)
	emailREST := mgtapi.Email(emailService)
	emailREST.Route(anon, bearer, basic)

	startupService := service.Startup(store, pusher)
	startupREST := mgtapi.Startup(startupService)
	startupREST.Route(anon, bearer, basic)

	sharedService := service.Shared()
	sharedAPI := mgtapi.Shared(sharedService)
	sharedAPI.Route(anon, bearer, basic)

	hardConfig := sonatype.HardConfig()
	sona := sonatype.NewClient(hardConfig, client)
	synchro := vulnsync.New(db, sona)
	mgtapi.Manual(synchro).Route(anon, bearer, basic)

	cmdb2Client := cmdb2.NewClient(cfg.Section.Cmdb2, client)
	cmdb2Service := service.Cmdb2(cmdb2Client)
	mgtapi.Cmdb2(cmdb2Service).Route(anon, bearer, basic)

	davREST := mgtapi.DavFS(base)
	davREST.Route(anon, bearer, basic)

	pprofDir := "resources/pprof"
	_ = os.RemoveAll(pprofDir)
	_ = os.MkdirAll(pprofDir, os.ModePerm)
	pprofService := service.Pprof(pprofDir, pusher)
	pprofREST := mgtapi.Pprof(pprofService)
	pprofREST.Route(anon, bearer, basic)

	app := &application{
		cfg:     cfg,
		handler: sh,
		parent:  ctx,
	}

	return app, nil
}
