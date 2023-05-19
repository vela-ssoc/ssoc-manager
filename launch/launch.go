package launch

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dbms"
	"github.com/vela-ssoc/vela-common-mb/logback"
	"github.com/vela-ssoc/vela-common-mb/problem"
	"github.com/vela-ssoc/vela-common-mb/taskpool"
	"github.com/vela-ssoc/vela-common-mb/validate"
	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-manager/app/mgtapi"
	"github.com/vela-ssoc/vela-manager/app/middle"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/app/session"
	"github.com/vela-ssoc/vela-manager/bridge/blink"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/infra/config"
	"github.com/vela-ssoc/vela-manager/infra/profile"
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
	logCfg := cfg.Logger

	zapLog := cfg.Logger.Zap()
	slog.Replace(zapLog)
	gormLog := logback.Gorm(zapLog, logCfg.Level)

	db, _, err := dbms.Open(dbCfg, gormLog)
	if err != nil {
		return nil, err
	}
	query.SetDefault(db)

	const name = "manager"
	const headerKey = ship.HeaderAuthorization
	queryKey := strings.ToLower(headerKey)
	auth := middle.Auth(headerKey, queryKey)

	secCfg := cfg.Section
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

	anon := sh.Group("/api/v1")
	bearer := anon.Clone().Use(auth.Bearer)
	basic := anon.Clone().Use(auth.Basic)

	// 初始化协程池
	pool := taskpool.NewPool(32, 128)

	// ==========[ broker begin ] ==========
	huber := linkhub.New(http.NewServeMux(), pool) // 将连接中心注入到 broker 接入网关中
	pusher := push.NewPush(huber)
	brkHandle := blink.New(huber)        // 将 broker 网关注入到 blink service 中
	blinkREST := mgtapi.Blink(brkHandle) // 构造 REST 层
	blinkREST.Route(anon, bearer, basic) // 注册路由用于调用
	if err = huber.RestDB(); err != nil {
		return nil, err
	}
	// ==========[ broker end ] ==========

	client := netutil.NewClient()
	_ = service.Emc(huber, client)

	digestService := service.Digest()
	sequenceService := service.Sequence()

	userService := service.User(digestService)
	userREST := mgtapi.User(userService)
	userREST.Route(anon, bearer, basic)

	verifyService := service.Verify(3, nil)              // 验证码 3 分钟有效期
	loginLockService := service.LoginLock(time.Hour, 10) // 每小时错误 10 次就锁定账户

	authService := service.Auth(verifyService, loginLockService, userService)
	authREST := mgtapi.Auth(authService)
	authREST.Route(anon, bearer, basic)

	minionService := service.Minion()
	minionREST := mgtapi.Minion(huber, minionService)
	minionREST.Route(anon, bearer, basic)

	intoService := service.Into(huber)
	intoREST := mgtapi.Into(intoService, headerKey, queryKey)
	intoREST.Route(anon, bearer, basic)

	tagService := service.Tag()
	tagREST := mgtapi.Tag(tagService)
	tagREST.Route(anon, bearer, basic)

	// -----[ 配置与发布 ]-----
	substanceService := service.Substance(huber, digestService)
	substanceREST := mgtapi.Substance(substanceService)
	substanceREST.Route(anon, bearer, basic)

	compoundService := service.Compound()
	compoundREST := mgtapi.Compound(compoundService)
	compoundREST.Route(anon, bearer, basic)

	effectService := service.Effect(pusher, sequenceService)
	effectREST := mgtapi.Effect(effectService)
	effectREST.Route(anon, bearer, basic)
	// -----[ 配置与发布 ]-----

	elasticService := service.Elastic(name)
	elasticREST := mgtapi.Elastic(elasticService, headerKey, queryKey)
	elasticREST.Route(anon, bearer, basic)

	oplogService := service.Oplog()
	oplogREST := mgtapi.Oplog(oplogService)
	oplogREST.Route(anon, bearer, basic)

	minionTaskService := service.MinionTask()
	minionTaskREST := mgtapi.MinionTask(minionTaskService)
	minionTaskREST.Route(anon, bearer, basic)

	minionLogonService := service.MinionLogon()
	minionLogonREST := mgtapi.MinionLogon(minionLogonService)
	minionLogonREST.Route(anon, bearer, basic)

	riskService := service.Risk()
	riskREST := mgtapi.Risk(riskService)
	riskREST.Route(anon, bearer, basic)

	eventService := service.Event()
	eventREST := mgtapi.Event(eventService)
	eventREST.Route(anon, bearer, basic)

	sbomService := service.SBOM()
	sbomREST := mgtapi.SBOM(sbomService)
	sbomREST.Route(anon, bearer, basic)

	cmdbService := service.Cmdb()
	cmdbREST := mgtapi.Cmdb(cmdbService)
	cmdbREST.Route(anon, bearer, basic)

	app := &application{
		cfg:     cfg,
		handler: sh,
		parent:  ctx,
	}

	return app, nil
}
