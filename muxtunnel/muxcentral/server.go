package muxcentral

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/vela-tunnel/muxconn"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Options struct {
	Validate   func(any) error
	Handler    http.Handler
	Loader     func(ctx context.Context, id bson.ObjectID) (*muxproto.BrokerBootConfig, error)
	Logger     *slog.Logger
	PerTimeout time.Duration
}

type server struct {
	db   polydb.Sourcer
	hub  muxproto.Huber
	opts Options
}

func NewServer(db polydb.Sourcer, hub muxproto.Huber, opts Options) any {
	return &server{
		db:   db,
		hub:  hub,
		opts: opts,
	}
}

//goland:noinspection GoUnhandledErrorResult
func (s *server) ServeMUX(mux muxconn.Muxer) error {
	defer mux.Close()

	now := time.Now()
	peer, sess, err := s.authentication(mux, now)
	if err != nil {
		return err
	}

	s.log().Debug("开始服务节点业务", "session", sess)
	err = s.serveHTTP(peer)
	s.log().Debug("节点下线", "session", sess, "err", err)
	s.disconnected(peer, sess)

	return nil
}

func (s *server) responseError(conn net.Conn, err *muxproto.AuthError) error {
	timeout := s.perTimeout()
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))

	return muxconn.WriteAuth(conn, err)
}

func (s *server) responseConfig(conn net.Conn, cfg *muxproto.BrokerBootConfig) error {
	timeout := s.perTimeout()
	resp := &muxproto.BrokerAuthResponse{Code: http.StatusAccepted, Config: cfg}
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))

	return muxconn.WriteAuth(conn, resp)
}

func (s *server) perTimeout() time.Duration {
	if d := s.opts.PerTimeout; d > 0 {
		return d
	}

	return 10 * time.Second
}

func (s *server) perContext() (context.Context, context.CancelFunc) {
	d := s.perTimeout()
	return context.WithTimeout(context.Background(), d)
}

func (s *server) serveHTTP(p muxproto.Peer) error {
	h := s.opts.Handler
	if h == nil {
		h = http.NotFoundHandler()
	}
	srv := &http.Server{
		Handler: h,
		BaseContext: func(net.Listener) context.Context {
			return muxproto.WithContext(context.Background(), p)
		},
	}
	mux := p.MUX()

	return srv.Serve(mux)
}

func (s *server) log() *slog.Logger {
	if log := s.opts.Logger; log != nil {
		return log
	}

	return slog.Default()
}

//goland:noinspection GoUnhandledErrorResult
func (s *server) authentication(mux muxconn.Muxer, now time.Time) (muxproto.Peer, *sessionData, error) {
	timeout := s.perTimeout()
	oncec := muxproto.NewOnceCloser(mux)
	timer := time.AfterFunc(timeout, oncec.Close) // 防止客户端建立连接后，但是不发起认证。
	conn, err := mux.Accept()
	timer.Stop()
	if err != nil {
		s.log().Info("等待认证通道超时", "err", err)
		return nil, nil, err
	}
	defer conn.Close()

	if oncec.Closed() {
		s.log().Info("等待认证通道超时")
		return nil, nil, net.ErrClosed
	}

	var auth muxproto.BrokerAuthRequest
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	if err = muxconn.ReadAuth(conn, &auth); err != nil {
		s.log().Warn("认证消息读取出错", "err", err)
		return nil, nil, err
	}
	if err = s.validAuthRequest(auth); err != nil {
		ae := &muxproto.AuthError{Code: http.StatusBadRequest, Text: err.Error()}
		s.log().Warn("认证消息参数校验出错", "err", ae)
		s.responseError(conn, ae)
		return nil, nil, ae
	}
	sess := &sessionData{Request: auth, ConnectedAt: now}

	brok, err := s.findBrokerBySecret(auth.Secret)
	if err != nil {
		ae := &muxproto.AuthError{Code: http.StatusNotFound, Text: err.Error()}
		s.log().Error("通过密钥查询节点错误", "request", auth, "err", ae)
		s.responseError(conn, ae)
		return nil, nil, ae
	}
	if brok.Status {
		ae := &muxproto.AuthError{Code: http.StatusConflict, Text: "节点已经在线（数据库校验）"}
		s.log().Warn("预检节点在线状态", "request", auth, "err", ae)
		s.responseError(conn, ae)

		return nil, nil, ae
	}

	bid := brok.ID
	peer := s.hub.Put(bid, mux, nil)
	if peer == nil {
		ae := &muxproto.AuthError{Code: http.StatusConflict, Text: "节点已经在线（连接池检查）"}
		s.log().Warn("预检节点在线状态", "request", auth, "err", ae)
		s.responseError(conn, ae)
		return nil, nil, ae
	}

	cfg, err := s.loadBrokerConfig(bid)
	if err != nil {
		s.hub.Del(bid) // 从连接池中删除

		ae := &muxproto.AuthError{Code: http.StatusBadRequest, Text: err.Error()}
		s.log().Warn("加载启动配置出错", "request", auth, "err", ae)
		s.responseError(conn, ae)
		return nil, nil, ae
	}
	if cfg == nil {
		s.log().Warn("加载启动配置为 NULL", "request", auth)
	}

	if err = s.responseConfig(conn, cfg); err != nil {
		s.hub.Del(bid) // 从连接池中删除

		ae := &muxproto.AuthError{Code: http.StatusInternalServerError, Text: err.Error()}
		s.log().Warn("回写验证通过报文出错", "request", auth, "err", ae)
		s.responseError(conn, ae)
		return nil, nil, ae
	}

	if err = s.updateBrokerOnline(peer, sess); err != nil {
		s.hub.Del(bid) // 从连接池中删除

		ae := &muxproto.AuthError{Code: http.StatusInternalServerError, Text: err.Error()}
		s.log().Error("修改数据库上线状态出错", "request", auth, "err", ae)
		s.responseError(conn, ae)
		return nil, nil, ae
	}

	return peer, sess, nil
}

func (s *server) disconnected(p muxproto.Peer, data *sessionData) {
	id, disconnAt := p.ID(), time.Now()

	mux := p.MUX()
	tx, rx := mux.Traffic() // 中心端侧 RX/TX 互换
	{
		filter := bson.D{{Key: "_id", Value: id}, {Key: "status", Value: true}}
		update := bson.M{"$set": bson.M{
			"status":                     false,
			"disconnected_at":            disconnAt,
			"tunnel_stat.receive_bytes":  rx,
			"tunnel_stat.transmit_bytes": tx,
		}}
		ctx, cancel := s.perContext()
		defer cancel()

		coll := s.db.MDB().Broker()
		ret, err := coll.UpdateOne(ctx, filter, update)
		if err != nil {
			s.log().Error("节点下线修改数据库出错", "request", data, "err", err)
		} else if ret.ModifiedCount <= 0 {
			s.log().Error("节点下线修改数据库未匹配到数据", "request", data)
		} else {
			s.log().Info("节点下线修改数据库完毕", "request", data)
		}
	}
	s.hub.Del(id) // 从 HUB 中删除

	{
		cumulative, _ := mux.NumStreams()
		data.TunnelStat.Cumulative = cumulative
		data.TunnelStat.ReceiveBytes = rx
		data.TunnelStat.TransmitBytes = tx
		his := &model.BrokerConnectHistory{
			BrokerID:       id,
			ExecuteStat:    data.ExecuteStat,
			TunnelStat:     data.TunnelStat,
			DisconnectedAt: disconnAt,
		}

		ctx, cancel := s.perContext()
		defer cancel()

		coll := s.db.MDB().BrokerConnectHistory()
		_, _ = coll.InsertOne(ctx, his)
	}

	s.log().Info("节点下线处理完毕", "request", data)
}

func (s *server) findBrokerBySecret(secret string) (*model.Broker, error) {
	ctx, cancel := s.perContext()
	defer cancel()

	coll := s.db.MDB().Broker()
	brok, err := coll.FindOne(ctx, bson.M{"secret": secret})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = errors.New("节点不存在") // 没有注册节点
		}
		return nil, err
	}

	return brok, nil
}

func (s *server) updateBrokerOnline(p muxproto.Peer, sess *sessionData) error {
	auth := sess.Request
	id, mux := p.ID(), p.MUX()
	library, module := mux.Library()
	tunstat := model.TunnelStat{
		Library:     library,
		Module:      module,
		Inet:        auth.Inet,
		ConnectedAt: sess.ConnectedAt,
		LocalAddr:   mux.RemoteAddr().String(),
		RemoteAddr:  mux.Addr().String(),
	}
	exestat := model.ExecuteStat{
		Goos:       auth.Goos,
		Goarch:     auth.Goarch,
		Version:    auth.Version,
		PID:        auth.PID,
		Args:       auth.Args,
		Hostname:   auth.Hostname,
		Workdir:    auth.Workdir,
		Executable: auth.Executable,
	}
	sess.TunnelStat = tunstat
	sess.ExecuteStat = exestat

	filter := bson.D{{Key: "_id", Value: id}, {Key: "status", Value: false}}
	update := bson.M{"$set": bson.M{
		"status":       true,
		"tunnel_stat":  tunstat,
		"execute_stat": exestat,
		"heartbeat_at": sess.ConnectedAt,
	}}

	ctx, cancel := s.perContext()
	defer cancel()

	coll := s.db.MDB().Broker()
	ret, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	} else if ret.ModifiedCount == 0 {
		return errors.New("没有修改任何数据")
	}

	return nil
}

func (s *server) validAuthRequest(req muxproto.BrokerAuthRequest) error {
	if f := s.opts.Validate; f != nil {
		return f(req)
	}

	var errs []error
	if req.Secret == "" {
		errs = append(errs, errors.New("连接密钥必须填写(secret)"))
	}
	if req.Version == "" {
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

func (s *server) loadBrokerConfig(id bson.ObjectID) (*muxproto.BrokerBootConfig, error) {
	f := s.opts.Loader
	if f == nil {
		return nil, nil
	}

	ctx, cancel := s.perContext()
	defer cancel()

	return f(ctx, id)
}
