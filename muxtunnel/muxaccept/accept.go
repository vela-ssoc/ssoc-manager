package muxaccept

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/vela-ssoc/ssoc-common/datalayer/model"
	"github.com/vela-ssoc/ssoc-common/datalayer/query"
	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-proto/muxconn"
	"github.com/vela-ssoc/ssoc-proto/muxproto"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

type Options struct {
	Huber      muxserver.Huber // 必须填写
	Handler    http.Handler
	Validator  func(any) error
	Logger     *slog.Logger
	PerTimeout time.Duration
	BootLoader muxserver.BootLoader[BrokConfig] // 必须填写，节点认证通过后加载启动配置。
	Notifier   muxserver.ConnectNotifier
}

func NewAccept(qry *query.Query, opts Options) muxserver.MUXAccepter {
	return &brokerAccept{
		qry:  qry,
		opts: opts,
	}
}

type brokerAccept struct {
	qry  *query.Query
	opts Options
}

//goland:noinspection GoUnhandledErrorResult
func (ms *brokerAccept) AcceptMUX(mux muxconn.Muxer) {
	defer mux.Close()

	connectAt := time.Now()
	attrs := []any{"connect_at", connectAt, "remote_addr", mux.RemoteAddr()}
	peer, err := ms.authentication(mux)
	if err != nil {
		attrs = append(attrs, "error", err)
		ms.log().Warn("节点上线失败", attrs...)

		if ntf := ms.opts.Notifier; ntf != nil {
			ctx, cancel := ms.perContext()
			defer cancel()

			ntf.OnAuthFailed(ctx, mux, connectAt, err)
		}

		return
	}

	id, info := peer.ID(), peer.Info()
	attrs = append(attrs, "id", id, "info", info)
	ms.log().Debug("节点上线成功", attrs...)

	if ntf := ms.opts.Notifier; ntf != nil {
		ctx, cancel := ms.perContext()
		defer cancel()

		ntf.OnConnected(ctx, info, connectAt)
	}

	ms.log().Debug("开始服务节点业务", attrs...)
	err = ms.serveHTTP(peer)
	attrs = append(attrs, "error", err)
	ms.log().Debug("节点下线", attrs...)

	ms.disconnected(peer, connectAt)
}

//goland:noinspection GoUnhandledErrorResult
func (ms *brokerAccept) authentication(mux muxconn.Muxer) (muxserver.Peer, error) {
	timeout := ms.perTimeout()
	oncec := muxserver.NewOnceCloser(mux)
	timer := time.AfterFunc(timeout, oncec.Close) // 防止客户端建立连接后，但是不发起认证。
	conn, err := mux.Accept()
	timer.Stop()
	if err != nil {
		ms.log().Info("等待认证通道超时", "error", err)
		return nil, err
	}
	defer conn.Close()

	if oncec.Closed() {
		ms.log().Info("等待认证通道超时")
		return nil, net.ErrClosed
	}

	req := new(AuthRequest)
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	if err = muxproto.ReadAuth(conn, req); err != nil {
		ms.log().Warn("认证消息读取出错", "error", err)
		return nil, err
	}
	if err = ms.validAuthRequest(req); err != nil {
		ae := &AuthError{Code: http.StatusBadRequest, Text: err.Error()}
		ms.log().Warn("认证消息参数校验出错", "error", ae)
		ms.responseError(conn, ae)
		return nil, ae
	}

	inf := muxserver.PeerInfo{
		Semver:   req.Semver,
		Inet:     req.Inet,
		Goos:     req.Goos,
		Goarch:   req.Goarch,
		Hostname: req.Hostname,
	}
	brok, err := ms.findBrokerBySecret(req.Secret)
	if err != nil {
		ae := &AuthError{Code: http.StatusNotFound, Text: err.Error()}
		ms.log().Error("通过密钥查询节点错误", "info", inf, "error", ae)
		ms.responseError(conn, ae)
		return nil, ae
	}

	id := brok.ID
	inf.Name = brok.Name
	attrs := []any{"info", inf}
	if brok.Status {
		ae := &AuthError{Code: http.StatusConflict, Text: "节点已经在线（数据库校验）"}
		attrs = append(attrs, "error", ae)
		ms.log().Warn("预检节点在线状态", attrs...)
		ms.responseError(conn, ae)

		return nil, ae
	}
	peer := ms.putHub(id, mux, inf)
	if peer == nil {
		ae := &AuthError{Code: http.StatusConflict, Text: "节点已经在线（连接池检查）"}
		attrs = append(attrs, "error", ae)
		ms.log().Warn("预检节点在线状态", attrs...)
		ms.responseError(conn, ae)
		return nil, ae
	}
	cfg, err := ms.loadConfig()
	if err != nil {
		ms.delHub(id) // 从连接池中删除

		ae := &AuthError{Code: http.StatusBadRequest, Text: err.Error()}
		attrs = append(attrs, "error", ae)
		ms.log().Warn("加载启动配置出错", attrs...)
		ms.responseError(conn, ae)
		return nil, ae
	}

	if err = ms.responseConfig(conn, cfg); err != nil {
		ms.delHub(id) // 从连接池中删除

		ae := &AuthError{Code: http.StatusInternalServerError, Text: err.Error()}
		attrs = append(attrs, "error", ae)
		ms.log().Warn("回写验证通过报文出错", attrs...)
		ms.responseError(conn, ae)
		return nil, ae
	}

	if err = ms.updateBrokerOnline(id, req); err != nil {
		ms.delHub(id) // 从连接池中删除

		ae := &AuthError{Code: http.StatusInternalServerError, Text: err.Error()}
		attrs = append(attrs, "error", ae)
		ms.log().Error("修改数据库上线状态出错", attrs...)
		ms.responseError(conn, ae)
		return nil, ae
	}

	return peer, nil
}

func (ms *brokerAccept) serveHTTP(peer muxserver.Peer) error {
	h := ms.opts.Handler
	if h == nil {
		h = http.NotFoundHandler()
	}
	srv := &http.Server{
		Handler: h,
		BaseContext: func(net.Listener) context.Context {
			return muxserver.WithContext(context.Background(), peer)
		},
	}
	mux := peer.MUX()

	return srv.Serve(mux)
}

func (ms *brokerAccept) disconnected(peer muxserver.Peer, connectAt time.Time) {
	disconnectAt := time.Now()
	id, info := peer.ID(), peer.Info()

	attrs := []any{"id", id, "info", info, "connect_at", connectAt, "disconnect_at", disconnectAt}
	// 修改数据库在线状态
	{
		ctx, cancel := ms.perContext()
		defer cancel()

		tbl := ms.qry.Broker
		dao := tbl.WithContext(ctx)

		wheres := []gen.Condition{tbl.ID.Eq(id), tbl.Status.Is(true)}
		ret, err := dao.Where(wheres...).UpdateSimple(tbl.Status.Value(false))
		if err != nil {
			attrs = append(attrs, "error", err)
			ms.log().Error("节点下线修改数据库出错", attrs...)
		} else if ret.RowsAffected == 0 {
			ms.log().Error("节点下线修改数据库未匹配到数据", attrs...)
		} else {
			ms.log().Info("节点下线修改数据库完毕", attrs...)
		}
	}
	ms.delHub(id) // 从 hub 中删除连接

	ms.log().Debug("通知节点下线事件", attrs...)
	if ntf := ms.opts.Notifier; ntf != nil {
		ctx, cancel := ms.perContext()
		defer cancel()

		ntf.OnDisconnected(ctx, info, connectAt, disconnectAt)
	}
}

func (ms *brokerAccept) perTimeout() time.Duration {
	if d := ms.opts.PerTimeout; d > 0 {
		return d
	}

	return 10 * time.Second
}

func (ms *brokerAccept) perContext() (context.Context, context.CancelFunc) {
	d := ms.perTimeout()

	return context.WithTimeout(context.Background(), d)
}

func (ms *brokerAccept) log() *slog.Logger {
	if l := ms.opts.Logger; l != nil {
		return l
	}

	return slog.Default()
}

func (ms *brokerAccept) putHub(id int64, mux muxconn.Muxer, inf muxserver.PeerInfo) muxserver.Peer {
	return ms.opts.Huber.Put(id, mux, inf)
}

func (ms *brokerAccept) delHub(id int64) {
	ms.opts.Huber.DelID(id)
}

func (ms *brokerAccept) validAuthRequest(req *AuthRequest) error {
	if v := ms.opts.Validator; v != nil {
		return v(req)
	}

	var errs []error
	if req.Secret == "" {
		errs = append(errs, errors.New("连接密钥必须填写(secret)"))
	}
	if req.Semver == "" {
		errs = append(errs, errors.New("版本号必须填写(semver)"))
	}
	if str := req.Inet; str == "" {
		errs = append(errs, errors.New("出口地址必须填写(inet)"))
	} else if _, err := netip.ParseAddr(str); err != nil {
		errs = append(errs, errors.New("出口地址格式错误(inet)"))
	}
	if req.Goos == "" {
		errs = append(errs, errors.New("操作系统类型必须填写(goos)"))
	}
	if req.Goarch == "" {
		errs = append(errs, errors.New("架构必须填写(goarch)"))
	}

	return errors.Join(errs...)
}

func (ms *brokerAccept) findBrokerBySecret(secret string) (*model.Broker, error) {
	ctx, cancel := ms.perContext()
	defer cancel()

	tbl := ms.qry.Broker
	dao := tbl.WithContext(ctx)

	brok, err := dao.Where(tbl.Secret.Eq(secret)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("未找到该节点") // 没有注册节点
		}
		return nil, err
	}

	return brok, nil
}

func (ms *brokerAccept) updateBrokerOnline(id int64, req *AuthRequest) error {
	tbl := ms.qry.Broker
	updates := []field.AssignExpr{
		tbl.Status.Value(true),
		tbl.Semver.Value(req.Semver),
		tbl.HeartbeatAt.Value(time.Now()),
	}

	ctx, cancel := ms.perContext()
	defer cancel()

	dao := tbl.WithContext(ctx)
	ret, err := dao.Where(tbl.ID.Eq(id), tbl.Status.Value(false)).
		UpdateSimple(updates...)
	if err != nil {
		return err
	} else if ret.RowsAffected == 0 {
		return errors.New("没有修改任何数据")
	}

	return nil
}

func (ms *brokerAccept) loadConfig() (*BrokConfig, error) {
	bl := ms.opts.BootLoader
	if bl == nil {
		return nil, errors.New("没有设置配置加载方式（ConfigLoader）")
	}

	ctx, cancel := ms.perContext()
	defer cancel()

	return bl.LoadBoot(ctx)
}

func (ms *brokerAccept) responseError(conn net.Conn, err *AuthError) error {
	timeout := ms.perTimeout()
	resp := &authResponse{Code: err.Code, Text: err.Text}
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))

	return muxproto.WriteAuth(conn, resp)
}

func (ms *brokerAccept) responseConfig(conn net.Conn, cfg *BrokConfig) error {
	timeout := ms.perTimeout()
	resp := &authResponse{Code: http.StatusAccepted, Config: cfg}
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))

	return muxproto.WriteAuth(conn, resp)
}
