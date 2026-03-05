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
	Notifier   Notifier
}

func NewAccept(db repository.Database, opts Options) muxserver.MUXAccepter {
	opts.Notifier = wrapSafeNotifier(opts.Notifier)

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
func (bs *brokerAccept) AcceptMUX(mux muxconn.Muxer) error {
	defer mux.Close()

	connectAt := time.Now()
	name, module := mux.Library()
	sessData := &tunnelSessionData{
		ConnectAt:     connectAt,
		LocalAddr:     mux.RemoteAddr().String(), // manager 记录 broker 视角，本地与远程地址互换
		RemoteAddr:    mux.Addr().String(),
		TunnelLibrary: model.TunnelLibrary{Name: name, Module: module},
	}

	peer, err := bs.authentication(mux, sessData)
	{
		ntf := bs.opts.Notifier
		ctx, cancel := bs.perContext()
		defer cancel()
		if err != nil {
			_ = ntf.OnFailed(ctx, mux, err)
			return err
		}

		id, info := peer.ID(), peer.Info()
		_ = ntf.OnConnected(ctx, id, info)
	}

	bs.log().Debug("开始服务节点业务", "session", sessData)
	err = bs.serveHTTP(peer)
	sessData.DisconnectAt = time.Now()
	bs.log().Debug("节点下线", "session", sessData, "error", err)

	bs.disconnected(sessData)

	return nil
}

func (bs *brokerAccept) AcceptTCP(w http.ResponseWriter, r *http.Request) error {
	return errors.ErrUnsupported
}

//goland:noinspection GoUnhandledErrorResult
func (bs *brokerAccept) authentication(mux muxconn.Muxer, sessData *tunnelSessionData) (muxserver.Peer, error) {
	timeout := bs.perTimeout()
	oncec := muxserver.NewOnceCloser(mux)
	timer := time.AfterFunc(timeout, oncec.Close) // 防止客户端建立连接后，但是不发起认证。
	conn, err := mux.Accept()
	timer.Stop()
	if err != nil {
		bs.log().Info("等待认证通道超时", "error", err)
		return nil, err
	}
	defer conn.Close()

	if oncec.Closed() {
		bs.log().Info("等待认证通道超时")
		return nil, net.ErrClosed
	}

	req := new(muxproto.BrokerAuthRequest)
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	if err = muxproto.ReadAuth(conn, req); err != nil {
		bs.log().Warn("认证消息读取出错", "error", err)
		return nil, err
	}
	if err = bs.validAuthRequest(req); err != nil {
		ae := &muxproto.AuthError{Code: http.StatusBadRequest, Text: err.Error()}
		bs.log().Warn("认证消息参数校验出错", "error", ae)
		bs.responseError(conn, ae)
		return nil, ae
	}

	inf := muxserver.PeerInfo{
		Semver:      req.Semver,
		Inet:        req.Inet,
		Goos:        req.Goos,
		Goarch:      req.Goarch,
		Hostname:    req.Hostname,
		ConnectedAt: sessData.ConnectAt,
	}
	brok, err := bs.findInstance(req.Secret)
	if err != nil {
		ae := &muxproto.AuthError{Code: http.StatusNotFound, Text: err.Error()}
		bs.log().Error("通过密钥查询节点错误", "info", inf, "error", ae)
		bs.responseError(conn, ae)
		return nil, ae
	}

	id, name := brok.ID, brok.Name
	inf.Instance = name
	sessData.ID = id
	sessData.Name = name
	sessData.Request = req

	if brok.Status {
		ae := &muxproto.AuthError{Code: http.StatusConflict, Text: "节点已经在线（数据库校验）"}
		bs.log().Warn("预检节点在线状态", "session", sessData, "error", ae)
		bs.responseError(conn, ae)

		return nil, ae
	}
	peer := bs.putHub(id, mux, inf)
	if peer == nil {
		ae := &muxproto.AuthError{Code: http.StatusConflict, Text: "节点已经在线（连接池检查）"}
		bs.log().Warn("预检节点在线状态", "session", sessData, "error", ae)
		bs.responseError(conn, ae)
		return nil, ae
	}

	sessData.Peer = peer
	cfg, err := bs.loadConfig()
	if err != nil {
		bs.delHub(id) // 从连接池中删除

		ae := &muxproto.AuthError{Code: http.StatusBadRequest, Text: err.Error()}
		bs.log().Warn("加载启动配置出错", "session", sessData, "error", ae)
		bs.responseError(conn, ae)
		return nil, ae
	}

	if err = bs.responseConfig(conn, cfg); err != nil {
		bs.delHub(id) // 从连接池中删除

		ae := &muxproto.AuthError{Code: http.StatusInternalServerError, Text: err.Error()}
		bs.log().Warn("回写验证通过报文出错", "session", sessData, "error", ae)
		bs.responseError(conn, ae)
		return nil, ae
	}

	if err = bs.updateOnline(sessData); err != nil {
		bs.delHub(id) // 从连接池中删除

		ae := &muxproto.AuthError{Code: http.StatusInternalServerError, Text: err.Error()}
		bs.log().Error("修改数据库上线状态出错", "session", sessData, "error", ae)
		bs.responseError(conn, ae)
		return nil, ae
	}

	return peer, nil
}

func (bs *brokerAccept) serveHTTP(peer muxserver.Peer) error {
	h := bs.opts.Handler
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

func (bs *brokerAccept) disconnected(sessData *tunnelSessionData) {
	id, peer := sessData.ID, sessData.Peer
	mux := peer.MUX()
	tx, rx := mux.Traffic()
	{
		filter := bson.D{{Key: "_id", Value: id}, {Key: "status", Value: true}}
		update := bson.M{"$set": bson.M{
			"status":                      false,
			"tunnel_stat.disconnected_at": sessData.DisconnectAt,
			"tunnel_stat.receive_bytes":   rx,
			"tunnel_stat.transmit_bytes":  tx,
		}}
		ctx, cancel := bs.perContext()
		defer cancel()

		coll := bs.db.Broker()
		ret, err := coll.UpdateOne(ctx, filter, update)
		if err != nil {
			bs.log().Error("节点下线修改数据库出错", "session", sessData, "error", err)
		} else if ret.ModifiedCount <= 0 {
			bs.log().Error("节点下线修改数据库未匹配到数据", "session", sessData)
		} else {
			bs.log().Info("节点下线修改数据库完毕", "session", sessData)
		}
	}
	bs.delHub(id) // 从 hub 中删除连接

	{
		cumulative, _ := mux.NumStreams()
		tunStat := sessData.TunnelStat
		tunStatHis := model.TunnelStatHistory{
			Inet:           tunStat.Inet,
			ConnectedAt:    sessData.ConnectAt,
			DisconnectedAt: sessData.DisconnectAt,
			ConnectSeconds: sessData.connectedSeconds(),
			Cumulative:     cumulative,
			Library:        tunStat.Library,
			LocalAddr:      sessData.LocalAddr,
			RemoteAddr:     sessData.RemoteAddr,
			ReceiveBytes:   rx,
			TransmitBytes:  tx,
		}
		his := &model.BrokerConnectHistory{
			BrokerID:    id,
			Name:        sessData.Name,
			ExecuteStat: sessData.ExecuteStat,
			TunnelStat:  tunStatHis,
		}

		ctx, cancel := bs.perContext()
		defer cancel()
		hisColl := bs.db.BrokerConnectHistory()
		_, _ = hisColl.InsertOne(ctx, his)
	}

	{
		ntf := bs.opts.Notifier
		ctx, cancel := bs.perContext()
		defer cancel()

		info := peer.Info()
		_ = ntf.OnDisconnected(ctx, id, info, sessData.DisconnectAt)
	}

	bs.log().Info("节点下线处理完毕", "session", sessData)
}

func (bs *brokerAccept) perTimeout() time.Duration {
	if d := bs.opts.PerTimeout; d > 0 {
		return d
	}

	return 10 * time.Second
}

func (bs *brokerAccept) perContext() (context.Context, context.CancelFunc) {
	d := bs.perTimeout()

	return context.WithTimeout(context.Background(), d)
}

func (bs *brokerAccept) log() *slog.Logger {
	if l := bs.opts.Logger; l != nil {
		return l
	}

	return slog.Default()
}

func (bs *brokerAccept) putHub(id bson.ObjectID, mux muxconn.Muxer, inf muxserver.PeerInfo) muxserver.Peer {
	return bs.opts.Huber.Put(id, mux, inf)
}

func (bs *brokerAccept) delHub(id bson.ObjectID) {
	bs.opts.Huber.DelID(id)
}

func (bs *brokerAccept) validAuthRequest(req *muxproto.BrokerAuthRequest) error {
	if v := bs.opts.Validator; v != nil {
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

func (bs *brokerAccept) findInstance(secret string) (*model.Broker, error) {
	ctx, cancel := bs.perContext()
	defer cancel()

	coll := bs.db.Broker()
	brok, err := coll.FindBySecret(ctx, secret)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = errors.New("节点不存在") // 没有注册节点
		}
		return nil, err
	}

	return brok, nil
}

func (bs *brokerAccept) updateOnline(sessData *tunnelSessionData) error {
	req, info := sessData.Request, sessData.Peer.Info()
	tunStat := model.TunnelStat{
		Inet:        info.Inet,
		ConnectedAt: sessData.ConnectAt,
		KeepaliveAt: sessData.ConnectAt,
		Library:     sessData.TunnelLibrary,
		LocalAddr:   sessData.LocalAddr,
		RemoteAddr:  sessData.RemoteAddr,
	}

	semver := req.Semver
	version := model.Semver(semver).Uint64()
	exeStat := model.ExecuteStat{
		Goos:       info.Goos,
		Goarch:     info.Goarch,
		Semver:     info.Semver,
		Version:    version,
		PID:        req.PID,
		Args:       req.Args,
		Hostname:   req.Hostname,
		Workdir:    req.Workdir,
		Executable: req.Executable,
	}
	sessData.ExecuteStat = exeStat
	sessData.TunnelStat = tunStat

	filter := bson.D{{Key: "_id", Value: sessData.ID}, {Key: "status", Value: false}}
	update := bson.M{"$set": bson.M{"status": true, "tunnel_stat": tunStat, "execute_stat": exeStat}}

	ctx, cancel := bs.perContext()
	defer cancel()

	coll := bs.db.Broker()
	ret, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	} else if ret.ModifiedCount == 0 {
		return errors.New("没有修改任何数据")
	}

	return nil
}

func (bs *brokerAccept) loadConfig() (*muxproto.BrokerBootConfig, error) {
	bl := bs.opts.BootLoader
	if bl == nil {
		return nil, errors.New("没有设置配置加载方式（ConfigLoader）")
	}

	ctx, cancel := bs.perContext()
	defer cancel()

	return bl.LoadBoot(ctx)
}

func (bs *brokerAccept) responseError(conn net.Conn, err *muxproto.AuthError) error {
	timeout := bs.perTimeout()
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))

	return muxproto.WriteAuth(conn, err)
}

func (bs *brokerAccept) responseConfig(conn net.Conn, cfg *muxproto.BrokerBootConfig) error {
	timeout := bs.perTimeout()
	resp := &muxproto.BrokerAuthResponse{Code: http.StatusAccepted, Config: cfg}
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))

	return muxproto.WriteAuth(conn, resp)
}
