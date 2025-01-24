package brkmux

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dbms"
	"github.com/vela-ssoc/vela-common-mba/smux"
	"github.com/vela-ssoc/vela-manager/infra/config"
	"github.com/vela-ssoc/vela-manager/profile"
	"github.com/xgfone/ship/v5"
	"golang.org/x/time/rate"
	"gorm.io/gen/field"
)

func New(qry *query.Query, api http.Handler, cfg *profile.Config, log *slog.Logger) *Mux {
	// 限流器防止 broker 瞬时大批量上线。
	// 由于 broker 总量一般在最多几十台，所以暂限流每秒 10 个。
	lim := rate.NewLimiter(10, 10)
	link := newLinker()

	return &Mux{
		qry:  qry,
		api:  api,
		cfg:  cfg,
		log:  log,
		link: link,
		lim:  lim,
	}
}

type Mux struct {
	// qry 数据库操作。
	qry *query.Query

	// api 当 broker 建立连接后，虚拟通道的 Handler。
	api http.Handler

	// cfg 全局配置。
	cfg *profile.Config

	// log 日志组件。
	log *slog.Logger

	// link 连接池
	link  *brokerLink
	lim   *rate.Limiter
	mutex sync.Mutex
	locks map[int64]struct{}
}

func (m *Mux) Linker() Linker {
	return m.link
}

func (m *Mux) Route(r *ship.RouteGroupBuilder) error {
	r.Route("/broker").
		POST(m.join).
		CONNECT(m.join)
	return nil
}

func (m *Mux) join(c *ship.Context) error {
	connectedAt := time.Now()
	ctx := c.Request().Context()

	clientIP := c.ClientIP()
	remoteAddr := c.Request().RemoteAddr
	attrs := []any{
		slog.String("client_ip", clientIP),
		slog.String("remote_addr", remoteAddr),
		slog.Time("connected_at", connectedAt),
	}

	if !m.lim.Allow() { // 限流器
		m.log.WarnContext(ctx, "broker 请求被限流器阻拦", attrs...)
		return nil
	}

	buf := make([]byte, 100*1024)
	n, _ := io.ReadFull(c.Body(), buf)
	slog.Int("data_length", n)
	m.log.InfoContext(ctx, "读取到协商数据", attrs...)

	var ident Ident
	if err := ident.decrypt(buf[:n]); err != nil {
		attrs = append(attrs, slog.Any("error", err))
		m.log.WarnContext(ctx, "协商数据解密失败", attrs...)
		return err
	}
	attrs = append(attrs, slog.Any("ident", ident))
	m.log.InfoContext(ctx, "协商数据解密成功", attrs...)

	// 查询节点认证信息
	tbl := m.qry.Broker
	dao := tbl.WithContext(ctx)
	brk, err := dao.Where(tbl.ID.Eq(ident.ID)).First()
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		m.log.WarnContext(ctx, "查询节点注册信息失败", attrs...)
		return err
	}
	if brk.Secret != ident.Secret {
		m.log.WarnContext(ctx, "密钥不匹配", attrs...)
		return nil
	}
	if brk.Status {
		m.log.WarnContext(ctx, "节点重复上线", attrs...)
		return nil
	}

	hijacker, ok := c.ResponseWriter().(http.Hijacker)
	if !ok {
		return nil
	}

	issue := Issue{
		Name:   brk.Name,
		Passwd: nil,
		Listen: Listen{
			Addr: brk.Bind,
		},
		Logger:   config.Logger{},
		Database: dbms.Config{},
		Section:  Section{},
	}
	enc, err := issue.encrypt()
	if err != nil {
		return err
	}
	reader := bytes.NewReader(enc)
	if err = c.Stream(http.StatusAccepted, ship.MIMETextPlain, reader); err != nil {
		return err
	}

	conn, _, err := hijacker.Hijack()
	if err != nil {
		return err
	}
	mux := smux.Server(conn, smux.DefaultConfig())
	bconn := &brokerConn{mux: mux, ident: ident, issue: issue, connectedAt: connectedAt}
	defer bconn.Close()

	m.link.putConn(bconn)
	defer m.link.delConn(ident.ID)

	srv := &http.Server{
		Handler: m.api,
		BaseContext: func(net.Listener) context.Context {
			return context.WithValue(context.Background(), brokerConnKey, bconn)
		},
	}

	if exx := srv.Serve(mux); exx != nil {
		attrs = append(attrs, slog.Any("smux_listen_error", exx))
	}

	disconnectedAt := time.Now()
	attrs = append(attrs, slog.Time("disconnected_at", disconnectedAt))
	m.log.WarnContext(ctx, "断开连接", attrs...)

	return nil
}

func (m *Mux) preflight(ctx context.Context, ident Ident) (*model.Broker, error) {
	tbl := m.qry.Broker
	brk, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(ident.ID)).
		First()
	if err != nil {
		m.log.WarnContext(ctx, "")
		return nil, err
	}
	if brk.Status {
		return nil, err
	}
	if brk.Secret != ident.Secret {
		return nil, err
	}

	return brk, nil
}

func (m *Mux) updateOnline(ctx context.Context, ident Ident) error {
	now := time.Now()
	tbl := m.qry.Broker
	columns := []field.AssignExpr{
		tbl.Status.Value(true),
		tbl.HeartbeatAt.Value(now),
		tbl.Semver.Value(ident.Semver),
	}
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(ident.ID), tbl.Status.Is(true)).
		UpdateSimple(columns...)

	return err
}

func (m *Mux) updateOffline(ctx context.Context, id int64) error {
	now := time.Now()
	tbl := m.qry.Broker
	columns := []field.AssignExpr{
		tbl.Status.Value(false),
		tbl.HeartbeatAt.Value(now),
	}
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id), tbl.Status.Is(true)).
		UpdateSimple(columns...)

	return err
}

func (m *Mux) lock(id int64) bool {
	m.mutex.Lock()
	_, exist := m.locks[id]
	if !exist {
		m.locks[id] = struct{}{}
	}
	m.mutex.Unlock()

	return !exist
}

func (m *Mux) unlock(id int64) {
	m.mutex.Lock()
	delete(m.locks, id)
	m.mutex.Unlock()
}
