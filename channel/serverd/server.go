package serverd

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/options"
	"github.com/vela-ssoc/ssoc-common/linkhub"
	"github.com/xtaci/smux"
	"gorm.io/gen/field"
)

type Handler interface {
	Handle(sess *smux.Session)
}

func New(qry *query.Query, bdc Database, opts ...options.Lister[option]) Handler {
	opts = append(opts, fallbackOption())
	opt := options.Eval(opts...)

	return &brokerServer{
		qry: qry,
		dbc: bdc,
		opt: opt,
	}
}

type brokerServer struct {
	qry *query.Query
	dbc Database
	opt option
}

func (as *brokerServer) Handle(sess *smux.Session) {
	defer sess.Close()

	timeout := as.opt.timeout
	peer, err := as.precheck(sess, timeout)
	if err != nil {
		as.log().Warn("节点上线认证失败", "error", err)
		return
	}
	defer as.disconnect(peer, timeout)

	as.opt.notifier.BrokerConnected(peer)

	srv := as.opt.server
	base := srv.BaseContext
	srv.BaseContext = func(ln net.Listener) context.Context {
		parent := context.Background()
		if base != nil {
			parent = base(ln)
		}
		return linkhub.WithContext(parent, peer)
	}

	lis := &smuxListener{sess: sess}
	err = srv.Serve(lis)

	as.log().Warn("broker 节点下线了", "error", err)
}

func (as *brokerServer) precheck(sess *smux.Session, timeout time.Duration) (linkhub.Peer, error) {
	timer := time.AfterFunc(timeout, func() { _ = sess.Close() })
	sig, err := sess.AcceptStream()
	timer.Stop()
	if err != nil {
		return nil, err
	}
	defer sig.Close()

	req := new(authRequest)
	_ = sig.SetDeadline(time.Now().Add(timeout))
	if err = linkhub.ReadAuth(sig, req); err != nil {
		return nil, err
	}

	// 参数校验
	if err = as.opt.valid(req); err != nil {
		resp := &authResponse{Code: http.StatusBadRequest, Message: err.Error()}
		_ = linkhub.WriteAuth(sig, resp)
		return nil, err
	}

	_, peer, code, err1 := as.join(sess, req, timeout)
	resp := &authResponse{Code: code}
	if err1 == nil {
		resp.Database = as.dbc
		if err2 := linkhub.WriteAuth(sig, resp); err2 != nil {
			return nil, err2
		}
		return peer, nil
	}

	resp.Message = err1.Error()
	_ = linkhub.WriteAuth(sig, resp)

	return nil, err1
}

func (as *brokerServer) join(sess *smux.Session, req *authRequest, timeout time.Duration) (*model.Broker, linkhub.Peer, int, error) {
	attrs := []any{slog.Any("broker_auth_request", req), slog.Duration("timeout", timeout)}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	tbl := as.qry.Broker
	dao := tbl.WithContext(ctx)

	secret := string(req.Secret)
	brk, err := dao.Where(tbl.Secret.Eq(secret)).First()
	if err != nil {
		return nil, nil, http.StatusNotFound, err
	}

	// 检查状态是否允许上线
	if brk.Status {
		as.log().Warn("broker 节点已经在线了（数据库检查）", attrs...)
		return brk, nil, http.StatusConflict, errors.New("节点重复上线")
	}

	brokerID := brk.ID
	peer := linkhub.NewPeer(brokerID, req.Semver, sess)
	if !as.opt.huber.Put(peer) {
		as.log().Warn("agent 节点已经在线了（内存检查）", attrs...)
		return brk, nil, http.StatusConflict, errors.New("节点重复上线")
	}

	// 修改数据库中的在线状态
	updates := []field.AssignExpr{
		tbl.Status.Value(true),
		tbl.Semver.Value(req.Semver),
		tbl.HeartbeatAt.Value(time.Now()),
	}

	if _, err = dao.Where(tbl.ID.Eq(brokerID), tbl.Status.Is(false)).
		UpdateSimple(updates...); err != nil {
		as.opt.huber.DelByID(brokerID)
		attrs = append(attrs, slog.Any("error", err))
		as.log().Warn("将数据库的节点状态标记为在线发生成错误", attrs...)
		return brk, nil, http.StatusInternalServerError, err
	}

	as.log().Info("broker 上线成功", attrs...)

	return brk, peer, http.StatusOK, nil
}

func (as *brokerServer) readRequest(stm *smux.Stream) (*authRequest, error) {
	head := make([]byte, 4)
	if n, err := io.ReadFull(stm, head); err != nil {
		return nil, err
	} else if n != 4 {
		return nil, io.ErrShortBuffer
	}

	size := binary.BigEndian.Uint32(head)
	data := make([]byte, size)
	if n, err := io.ReadFull(stm, data); err != nil {
		return nil, err
	} else if n != int(size) {
		return nil, io.ErrShortBuffer
	}

	req := new(authRequest)
	err := json.Unmarshal(data, req)

	return req, err
}

func (as *brokerServer) disconnect(peer linkhub.Peer, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	id := peer.Info().ID
	tbl := as.qry.Broker
	dao := tbl.WithContext(ctx)
	if _, err := dao.Where(tbl.ID.Eq(id), tbl.Status.Is(true)).
		UpdateSimple(tbl.Status.Value(false)); err != nil {
		as.log().Warn("修改 broker 节点下线状态失败", "error", err)
	}
	as.opt.huber.DelByID(id)
	as.opt.notifier.BrokerDisconnected(id)
}

func (as *brokerServer) log() *slog.Logger {
	if l := as.opt.logger; l != nil {
		return l
	}

	return slog.Default()
}

type smuxListener struct {
	sess *smux.Session
}

func (sl *smuxListener) Accept() (net.Conn, error) {
	stm, err := sl.sess.AcceptStream()
	if err != nil {
		return nil, err
	}

	return stm, nil
}

func (sl *smuxListener) Close() error {
	return sl.sess.Close()
}

func (sl *smuxListener) Addr() net.Addr {
	return sl.sess.LocalAddr()
}
