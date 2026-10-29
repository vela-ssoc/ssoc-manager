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
	"github.com/vela-ssoc/ssoc-common-mb/pubsub"
	"github.com/vela-ssoc/ssoc-manager/config"
	"github.com/vela-ssoc/vela-tunnel/muxconn"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Options struct {
	Validate   func(any) error
	Handler    http.Handler
	Logger     *slog.Logger
	PerTimeout time.Duration
}

type server struct {
	db   polydb.Sourcer
	hub  muxproto.Huber
	boot config.Config
	cfg  *model.ManagerConfig
	pub  pubsub.Publisher
	opts Options
}

func NewServer(db polydb.Sourcer, hub muxproto.Huber, boot config.Config, cfg *model.ManagerConfig, pub pubsub.Publisher, opts Options) muxproto.MUXServer {
	return &server{
		db:   db,
		hub:  hub,
		boot: boot,
		cfg:  cfg,
		pub:  pub,
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

	evtdata := EventData{
		Slug:        sess.Model.Slug,
		Name:        sess.Model.Name,
		Inet:        sess.Request.Inet,
		Version:     sess.Request.Version,
		ConnectedAt: now,
	}
	s.pub.Publish("broker.online", evtdata)

	err = s.serveHTTP(peer)
	disAt := time.Now()

	s.log().Debug("节点下线", "session", sess, "err", err)
	s.disconnected(peer, sess, disAt)

	evtdata.DisconnectedAt = disAt
	s.pub.Publish("broker.offline", evtdata)

	return nil
}

func (s *server) responseError(conn net.Conn, err *muxproto.AuthError) error {
	timeout := s.perTimeout()
	_ = conn.SetWriteDeadline(time.Now().Add(timeout))

	return muxconn.WriteAuth(conn, err)
}

func (s *server) responseConfig(conn net.Conn, cfg model.BrokerConfig) error {
	timeout := s.perTimeout()
	resp := muxproto.NewAcceptedBrokerAuthResponse(cfg)
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

func (s *server) serveHTTP(p muxproto.Node) error {
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
func (s *server) authentication(mux muxconn.Muxer, now time.Time) (muxproto.Node, *sessionData, error) {
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
	sess.Model = brok
	event := EventData{
		Slug:        brok.Slug,
		Name:        brok.Name,
		Inet:        auth.Inet,
		Version:     auth.Version,
		ConnectedAt: now,
	}

	if brok.Status {
		msg := "节点已经在线（数据库校验）"
		ae := &muxproto.AuthError{Code: http.StatusConflict, Text: msg}
		s.log().Warn("预检节点在线状态", "request", auth, "err", ae)
		s.responseError(conn, ae)

		event.Text = msg
		s.pub.Publish("broker.conflict", event)

		return nil, nil, ae
	}

	slug := brok.Slug
	peer := s.hub.Put(slug, mux, nil)
	if peer == nil {
		msg := "节点已经在线（连接池检查）"
		ae := &muxproto.AuthError{Code: http.StatusConflict, Text: msg}
		s.log().Warn("预检节点在线状态", "request", auth, "err", ae)
		s.responseError(conn, ae)

		event.Text = msg
		s.pub.Publish("broker.conflict", event)

		return nil, nil, ae
	}

	cfg := s.buildBrokerConfig(brok)
	if err = s.responseConfig(conn, cfg); err != nil {
		s.hub.Del(slug) // 从连接池中删除

		ae := &muxproto.AuthError{Code: http.StatusInternalServerError, Text: err.Error()}
		s.log().Warn("回写验证通过报文出错", "request", auth, "err", ae)
		s.responseError(conn, ae)
		return nil, nil, ae
	}

	if err = s.updateBrokerOnline(peer, sess); err != nil {
		s.hub.Del(slug) // 从连接池中删除

		ae := &muxproto.AuthError{Code: http.StatusInternalServerError, Text: err.Error()}
		s.log().Error("修改数据库上线状态出错", "request", auth, "err", ae)
		s.responseError(conn, ae)

		return nil, nil, ae
	}

	return peer, sess, nil
}

func (s *server) disconnected(p muxproto.Node, data *sessionData, disconnAt time.Time) {
	slug, mux := data.Model.Slug, p.MUX()

	tx, rx := mux.Traffic() // 中心端侧 RX/TX 互换
	{
		filter := bson.D{{Key: "slug", Value: slug}, {Key: "status", Value: true}}
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
	s.hub.Del(slug) // 从 HUB 中删除

	{
		sec := int64(disconnAt.Sub(data.ConnectedAt).Seconds())
		cumulative, _ := mux.NumStreams()
		data.TunnelStat.Cumulative = cumulative
		data.TunnelStat.ReceiveBytes = rx
		data.TunnelStat.TransmitBytes = tx
		his := &model.BrokerConnectHistory{
			BrokerID:       data.Model.ID,
			Slug:           slug,
			ExecuteStat:    data.ExecuteStat,
			TunnelStat:     data.TunnelStat,
			Seconds:        sec,
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

func (s *server) updateBrokerOnline(p muxproto.Node, sess *sessionData) error {
	auth := sess.Request
	slug, mux := p.Host(), p.MUX()
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

	filter := bson.D{{Key: "slug", Value: slug}, {Key: "status", Value: false}}
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

func (s *server) buildBrokerConfig(brok *model.Broker) model.BrokerConfig {
	sc, cfg := s.cfg, brok.Config
	cfg.Slug = brok.Slug
	if cfg.Server == nil {
		cfg.Server = &model.BrokerServerConfig{Addr: sc.Server.Addr}
	}
	if cfg.Database == nil {
		cfg.Database = new(model.BrokerDatabaseConfig)
	}
	if cfg.Database.URI == "" {
		cfg.Database.URI = s.boot.Database.URI
	}
	if cfg.Database.DSN == "" {
		cfg.Database.DSN = s.boot.Database.DSN
	}
	if cfg.Logger == nil {
		l := sc.Logger
		cfg.Logger = &model.LoggerConfig{
			Level:      l.Level,
			Console:    l.Console,
			Filename:   l.Filename,
			MaxSize:    l.MaxSize,
			MaxAge:     l.MaxAge,
			MaxBackups: l.MaxBackups,
			LocalTime:  l.LocalTime,
			Compress:   l.Compress,
		}
	}

	return cfg
}
