package muxaccept

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-common/store/model"
	"github.com/vela-ssoc/ssoc-common/store/repository"
	"github.com/vela-ssoc/ssoc-proto/muxconn"
	"github.com/vela-ssoc/ssoc-proto/muxproto"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Options struct {
	Huber      muxserver.Huber // 必须填写
	Handler    http.Handler
	Validator  func(any) error
	Logger     *slog.Logger
	PerTimeout time.Duration
	BootLoader muxserver.BootLoader[muxproto.BrokerBootConfig] // 必须填写，节点认证通过后加载启动配置。
	Notifier   muxserver.ConnectNotifier
}

func NewAccept(db repository.Database, opts Options) muxserver.MUXAccepter {
	return &brokerAccept{
		db:   db,
		opts: opts,
	}
}

type brokerAccept struct {
	db   repository.Database
	opts Options
}

//goland:noinspection GoUnhandledErrorResult
func (ms *brokerAccept) AcceptMUX(mux muxconn.Muxer) {
	defer mux.Close()

	connectAt := time.Now()
	attrs := []any{"connect_at", connectAt, "remote_addr", mux.RemoteAddr()}
	peer, err := ms.authentication(mux, connectAt)
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

	ms.disconnected(peer)
}

//goland:noinspection GoUnhandledErrorResult
func (ms *brokerAccept) authentication(mux muxconn.Muxer, connectAt time.Time) (muxserver.Peer, error) {
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

	req := new(muxproto.BrokerAuthRequest)
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	if err = muxproto.ReadAuth(conn, req); err != nil {
		ms.log().Warn("认证消息读取出错", "error", err)
		return nil, err
	}
	if err = ms.validAuthRequest(req); err != nil {
		ae := &muxproto.AuthError{Code: http.StatusBadRequest, Text: err.Error()}
		ms.log().Warn("认证消息参数校验出错", "error", ae)
		ms.responseError(conn, ae)
		return nil, ae
	}

	inf := muxserver.PeerInfo{
		Semver:      req.Semver,
		Inet:        req.Inet,
		Goos:        req.Goos,
		Goarch:      req.Goarch,
		Hostname:    req.Hostname,
		ConnectedAt: connectAt,
	}
	brok, err := ms.lookupBroker(req.Secret)
	if err != nil {
		ae := &muxproto.AuthError{Code: http.StatusNotFound, Text: err.Error()}
		ms.log().Error("通过密钥查询节点错误", "info", inf, "error", ae)
		ms.responseError(conn, ae)
		return nil, ae
	}

	id := brok.ID
	inf.Name = brok.Name
	attrs := []any{"info", inf}
	if brok.Status {
		ae := &muxproto.AuthError{Code: http.StatusConflict, Text: "节点已经在线（数据库校验）"}
		attrs = append(attrs, "error", ae)
		ms.log().Warn("预检节点在线状态", attrs...)
		ms.responseError(conn, ae)

		return nil, ae
	}
	peer := ms.putHub(id, mux, inf)
	if peer == nil {
		ae := &muxproto.AuthError{Code: http.StatusConflict, Text: "节点已经在线（连接池检查）"}
		attrs = append(attrs, "error", ae)
		ms.log().Warn("预检节点在线状态", attrs...)
		ms.responseError(conn, ae)
		return nil, ae
	}
	cfg, err := ms.loadConfig()
	if err != nil {
		ms.delHub(id) // 从连接池中删除

		ae := &muxproto.AuthError{Code: http.StatusBadRequest, Text: err.Error()}
		attrs = append(attrs, "error", ae)
		ms.log().Warn("加载启动配置出错", attrs...)
		ms.responseError(conn, ae)
		return nil, ae
	}

	if err = ms.responseConfig(conn, cfg); err != nil {
		ms.delHub(id) // 从连接池中删除

		ae := &muxproto.AuthError{Code: http.StatusInternalServerError, Text: err.Error()}
		attrs = append(attrs, "error", ae)
		ms.log().Warn("回写验证通过报文出错", attrs...)
		ms.responseError(conn, ae)
		return nil, ae
	}

	if err = ms.updateBrokerOnline(id, mux, req); err != nil {
		ms.delHub(id) // 从连接池中删除

		ae := &muxproto.AuthError{Code: http.StatusInternalServerError, Text: err.Error()}
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

func (ms *brokerAccept) disconnected(peer muxserver.Peer) {
	disconnectAt := time.Now()
	id, info := peer.ID(), peer.Info()
	mux := peer.MUX()

	attrs := []any{"id", id, "info", info, "connect_at", info.ConnectedAt, "disconnect_at", disconnectAt}
	// 修改数据库在线状态
	{
		tx, rx := mux.Traffic() // 站在 manager 视角记录 broker，RX TX 互换。
		filter := bson.D{{Key: "_id", Value: id}, {Key: "status", Value: true}}
		update := bson.M{"$set": bson.M{
			"status":                      false,
			"tunnel_stat.disconnected_at": disconnectAt,
			"tunnel_stat.receive_bytes":   rx,
			"tunnel_stat.transmit_bytes":  tx,
		}}
		ctx, cancel := ms.perContext()
		defer cancel()

		coll := ms.db.Broker()
		ret, err := coll.UpdateOne(ctx, filter, update)
		if err != nil {
			attrs = append(attrs, "error", err)
			ms.log().Error("节点下线修改数据库出错", attrs...)
		} else if ret.ModifiedCount <= 0 {
			ms.log().Error("节点下线修改数据库未匹配到数据", attrs...)
		} else {
			ms.log().Info("节点下线修改数据库完毕", attrs...)
		}
	}
	ms.delHub(id) // 从 hub 中删除连接

	tx, rx := mux.Traffic()
	name, module := mux.Library()
	connectAt := info.ConnectedAt
	raddr, laddr := mux.Addr(), mux.RemoteAddr()
	his := &model.BrokerConnectHistory{
		BrokerID: id,
		Name:     info.Name,
		Semver:   info.Semver,
		Inet:     info.Inet,
		Goos:     info.Goos,
		Goarch:   info.Goarch,
		TunnelStat: model.TunnelStatHistory{
			ConnectedAt:    connectAt,
			DisconnectedAt: disconnectAt,
			Library:        model.TunnelLibrary{Name: name, Module: module},
			LocalAddr:      laddr.String(),
			RemoteAddr:     raddr.String(),
			ReceiveBytes:   rx,
			TransmitBytes:  tx,
		},
	}
	{
		ctx, cancel := ms.perContext()
		defer cancel()
		hisColl := ms.db.BrokerConnectHistory()
		_, _ = hisColl.InsertOne(ctx, his)
	}

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

func (ms *brokerAccept) putHub(id bson.ObjectID, mux muxconn.Muxer, inf muxserver.PeerInfo) muxserver.Peer {
	return ms.opts.Huber.Put(id, mux, inf)
}

func (ms *brokerAccept) delHub(id bson.ObjectID) {
	ms.opts.Huber.DelID(id)
}

func (ms *brokerAccept) validAuthRequest(req *muxproto.BrokerAuthRequest) error {
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

func (ms *brokerAccept) lookupBroker(secret string) (*model.Broker, error) {
	ctx, cancel := ms.perContext()
	defer cancel()

	coll := ms.db.Broker()
	brok, err := coll.FindBySecret(ctx, secret)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = errors.New("节点不存在") // 没有注册节点
		}
		return nil, err
	}

	return brok, nil
}

func (ms *brokerAccept) updateBrokerOnline(id bson.ObjectID, mux muxconn.Muxer, req *muxproto.BrokerAuthRequest) error {
	now := time.Now()
	name, module := mux.Library()
	laddr, raddr := mux.RemoteAddr(), mux.Addr() // manager 记录 broker 视角，本地与远程地址互换

	tunStat := &model.TunnelStat{
		ConnectedAt: now,
		KeepaliveAt: now,
		Library:     model.TunnelLibrary{Name: name, Module: module},
		LocalAddr:   laddr.String(),
		RemoteAddr:  raddr.String(),
	}
	exeStat := &model.ExecuteStat{
		Inet:       req.Inet,
		Goos:       req.Goos,
		Goarch:     req.Goarch,
		Semver:     req.Semver,
		PID:        req.PID,
		Args:       req.Args,
		Hostname:   req.Hostname,
		Workdir:    req.Workdir,
		Executable: req.Executable,
	}

	filter := bson.D{{Key: "_id", Value: id}, {Key: "status", Value: false}}
	update := bson.M{"$set": bson.M{"status": true, "tunnel_stat": tunStat, "execute_stat": exeStat}}

	ctx, cancel := ms.perContext()
	defer cancel()

	coll := ms.db.Broker()
	ret, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	} else if ret.ModifiedCount == 0 {
		return errors.New("没有修改任何数据")
	}

	return nil
}

func (ms *brokerAccept) loadConfig() (*muxproto.BrokerBootConfig, error) {
	bl := ms.opts.BootLoader
	if bl == nil {
		return nil, errors.New("没有设置配置加载方式（ConfigLoader）")
	}

	ctx, cancel := ms.perContext()
	defer cancel()

	return bl.LoadBoot(ctx)
}

func (ms *brokerAccept) responseError(conn net.Conn, err *muxproto.AuthError) error {
	timeout := ms.perTimeout()
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))

	return muxproto.WriteAuth(conn, err)
}

func (ms *brokerAccept) responseConfig(conn net.Conn, cfg *muxproto.BrokerBootConfig) error {
	timeout := ms.perTimeout()
	resp := &muxproto.BrokerAuthResponse{Code: http.StatusAccepted, Config: cfg}
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))

	return muxproto.WriteAuth(conn, resp)
}
