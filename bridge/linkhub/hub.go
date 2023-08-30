package linkhub

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/gopool"
	"github.com/vela-ssoc/vela-common-mb/problem"
	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-common-mba/smux"
	"github.com/vela-ssoc/vela-manager/bridge/blink"
	"github.com/vela-ssoc/vela-manager/infra/config"
)

var (
	ErrBrokerNotFound = errors.New("broker 节点不存在")
	ErrBrokerRepeat   = errors.New("broker 节点重复连接")
	ErrBrokerInet     = errors.New("broker IP 不合法")
	ErrBrokerOffline  = errors.New("代理节点未上线")
)

type Huber interface {
	blink.Joiner

	ResetDB() error

	Do(ctx context.Context, bid int64, method, path string, body io.Reader, header http.Header) (*http.Response, error)

	Oneway(ctx context.Context, id int64, path string, req any) error

	Unicast(ctx context.Context, id int64, path string, req, resp any) error

	Multicast(ctx context.Context, bids []int64, path string, req any) <-chan *ErrorFuture

	Broadcast(ctx context.Context, path string, req any) <-chan *ErrorFuture

	// Stream 向 broker 节点建立 websocket 连接
	Stream(ctx context.Context, bid int64, path string, header http.Header) (*websocket.Conn, *http.Response, error)

	// Forward 向 broker 节点转发 http 请求
	Forward(bid int64, w http.ResponseWriter, r *http.Request)
}

func New(handler http.Handler, pool gopool.Executor, cfg config.Config) Huber {
	hub := &brokerHub{
		name:     "manager",
		handler:  handler,
		config:   cfg,
		client:   netutil.HTTPClient{},
		connects: make(map[string]*spdyServerConn, 16),
		random:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	trip := &http.Transport{DialContext: hub.dialContext}
	hub.client = netutil.NewClient(trip)
	hub.forward = netutil.NewForward(trip, hub.errorFunc)
	hub.streamer = netutil.NewStream(hub.dialContext)
	hub.pool = pool

	return hub
}

type brokerHub struct {
	name     string
	handler  http.Handler
	config   config.Config
	client   netutil.HTTPClient
	forward  netutil.Forwarder
	streamer netutil.Streamer
	pool     gopool.Executor
	mutex    sync.RWMutex
	connects map[string]*spdyServerConn
	random   *rand.Rand
}

func (hub *brokerHub) Name() string {
	return hub.name
}

func (hub *brokerHub) Auth(ctx context.Context, ident blink.Ident) (blink.Issue, http.Header, error) {
	var issue blink.Issue
	id, secret, inet := ident.ID, ident.Secret, ident.Inet
	if len(inet) == 0 || inet.IsLoopback() {
		return issue, nil, ErrBrokerInet
	}

	// 查询 broker
	brkTbl := query.Broker
	brk, err := brkTbl.WithContext(ctx).
		Where(brkTbl.ID.Eq(id), brkTbl.Secret.Eq(secret)).
		First()
	if err != nil {
		return issue, nil, ErrBrokerNotFound
	}

	sid := strconv.FormatInt(id, 10)
	if brk.Status || hub.getConn(sid) != nil {
		return issue, nil, ErrBrokerRepeat
	}

	// 随机生成一个 32-64 位的加密密钥
	psz := hub.random.Intn(33) + 32
	passwd := make([]byte, psz)
	_, _ = hub.random.Read(passwd)

	issue.Name, issue.Passwd = brk.Name, passwd
	issue.Listen = blink.Listen{Addr: brk.Bind}
	issue.Logger, issue.Database = hub.config.Logger, hub.config.Database

	return issue, nil, nil
}

func (hub *brokerHub) Join(tran net.Conn, ident blink.Ident, issue blink.Issue) error {
	conn := hub.newConn(tran, ident, issue)
	if !hub.putConn(conn) {
		return ErrBrokerRepeat
	}
	sid := conn.sid
	defer hub.delConn(sid)

	tbl := query.Broker
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	info, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(ident.ID), tbl.Status.Is(false)).
		UpdateColumn(tbl.Status, true)
	cancel()
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return ErrBrokerOffline
	}

	defer func() {
		ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
		_, _ = tbl.WithContext(ctx).
			Where(tbl.ID.Eq(ident.ID)).
			Where(tbl.Status.Is(true)).
			UpdateColumn(tbl.Status, false)
		cancel()
	}()

	srv := &http.Server{
		Handler: hub.handler,
		BaseContext: func(net.Listener) context.Context {
			return context.WithValue(context.Background(), brokerCtxKey, conn)
		},
	}

	return srv.Serve(conn.muxer) // 此处会阻塞，一旦执行结束说明连接断开
}

func (hub *brokerHub) ResetDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	brk := query.Broker
	_, err := brk.WithContext(ctx).
		Where(brk.Status.Is(true)).
		UpdateColumn(brk.Status, false)
	return err
}

func (hub *brokerHub) Do(ctx context.Context, bid int64, method, path string, body io.Reader, header http.Header) (*http.Response, error) {
	httpURL := hub.httpURL(bid, path)
	req, err := http.NewRequestWithContext(ctx, method, httpURL, body)
	if err != nil {
		return nil, err
	}
	if len(header) > 0 {
		req.Header = header
	}

	return hub.client.Do(req)
}

func (hub *brokerHub) Oneway(ctx context.Context, id int64, path string, req any) error {
	wg := new(sync.WaitGroup)
	wg.Add(1)

	tsk := &onewayTask{
		wg:   wg,
		hub:  hub,
		bid:  id,
		path: path,
		req:  req,
	}
	hub.pool.Submit(tsk) // 交给协程池去执行

	return tsk.Wait()
}

func (hub *brokerHub) Unicast(ctx context.Context, id int64, path string, req, resp any) error {
	wg := new(sync.WaitGroup)
	wg.Add(1)

	rt := &resultTask{
		wg:    wg,
		huber: hub,
		id:    id,
		path:  path,
		req:   req,
		resp:  resp,
	}

	hub.pool.Submit(rt) // 交给协程池去执行
	err := rt.Wait()    // 等待执行完毕

	return err
}

func (hub *brokerHub) Multicast(ctx context.Context, bids []int64, path string, req any) <-chan *ErrorFuture {
	size := len(bids)
	ret := make(chan *ErrorFuture, size)
	if size == 0 {
		close(ret)
	} else {
		go hub.multicast(bids, path, req, ret)
	}

	return ret
}

func (hub *brokerHub) Broadcast(ctx context.Context, path string, req any) <-chan *ErrorFuture {
	bids := hub.keys() // 获取在线的 broker
	return hub.Multicast(ctx, bids, path, req)
}

func (hub *brokerHub) Forward(bid int64, w http.ResponseWriter, r *http.Request) {
	host := strconv.FormatInt(bid, 10)
	r.URL.Host = host
	r.URL.Scheme = "http"
	hub.forward.Forward(w, r)
}

func (hub *brokerHub) Stream(ctx context.Context, id int64, path string, header http.Header) (*websocket.Conn, *http.Response, error) {
	addr := hub.wsURL(id, path)
	return hub.streamer.Stream(ctx, addr, header)
}

func (hub *brokerHub) multicast(bids []int64, path string, req any, ret chan *ErrorFuture) {
	wg := new(sync.WaitGroup)
	for _, bid := range bids {
		tsk := &silentTask{
			wg:   wg,
			ret:  ret,
			hub:  hub,
			bid:  bid,
			path: path,
			req:  req,
		}

		wg.Add(1)
		hub.pool.Submit(tsk) // 将任务丢到协程池中等待执行
	}

	// 当所有的任务执行完毕，关闭返回值 chan
	wg.Wait()
	close(ret)
}

// silentJSON 发送 JSON 请求但不关心返回的 Body，只关注是否有错。
func (hub *brokerHub) silentJSON(ctx context.Context, id int64, path string, req any) error {
	addr := hub.httpURL(id, path)

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	return hub.client.SilentJSON(nil, http.MethodPost, addr, req, nil)
}

// sendJSON 发送 JSON 请求响应 JSON 数据
func (hub *brokerHub) sendJSON(ctx context.Context, id int64, path string, req, resp any) error {
	addr := hub.httpURL(id, path)

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	return hub.client.JSON(ctx, http.MethodPost, addr, req, resp, nil)
}

func (hub *brokerHub) newConn(tran net.Conn, ident blink.Ident, issue blink.Issue) *spdyServerConn {
	id := ident.ID
	sid := strconv.FormatInt(id, 10)
	cfg := smux.DefaultConfig()
	cfg.KeepAliveDisabled = true
	muxer := smux.Server(tran, cfg)

	return &spdyServerConn{
		id:    id,
		sid:   sid,
		muxer: muxer,
		ident: ident,
		issue: issue,
	}
}

func (hub *brokerHub) httpURL(id int64, path string) string {
	return hub.newURL(id, "http", path)
}

func (hub *brokerHub) wsURL(id int64, path string) string {
	return hub.newURL(id, "ws", path)
}

func (*brokerHub) newURL(id int64, scheme, path string) string {
	sid := strconv.FormatInt(id, 10)
	sn := strings.SplitN(path, "?", 2)
	u := &url.URL{Scheme: scheme, Host: sid, Path: sn[0]}
	if len(sn) == 2 {
		u.RawQuery = sn[1]
	}
	return u.String()
}

// getConn 通过 ID 获取连接
func (hub *brokerHub) getConn(id string) *spdyServerConn {
	hub.mutex.RLock()
	conn := hub.connects[id]
	hub.mutex.RUnlock()

	return conn
}

func (hub *brokerHub) putConn(conn *spdyServerConn) bool {
	sid := conn.sid
	hub.mutex.Lock()
	_, ok := hub.connects[sid]
	if !ok {
		hub.connects[sid] = conn
	}
	hub.mutex.Unlock()

	return !ok
}

func (hub *brokerHub) delConn(id string) {
	hub.mutex.Lock()
	delete(hub.connects, id)
	hub.mutex.Unlock()
}

func (hub *brokerHub) dialContext(_ context.Context, _, addr string) (net.Conn, error) {
	id, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, net.InvalidAddrError(addr)
	}

	if conn := hub.getConn(id); conn != nil {
		if stream, exx := conn.muxer.OpenStream(); exx != nil {
			return nil, exx
		} else {
			return stream, nil
		}
	}

	return nil, ErrBrokerOffline
}

func (hub *brokerHub) errorFunc(w http.ResponseWriter, r *http.Request, err error) {
	pd := &problem.Detail{
		Type:     hub.name,
		Title:    "网关错误",
		Status:   http.StatusBadGateway,
		Detail:   err.Error(),
		Instance: r.RequestURI,
	}
	_ = pd.JSON(w)
}

func (hub *brokerHub) keys() []int64 {
	hub.mutex.RLock()
	ret := make([]int64, 0, len(hub.connects))
	for _, conn := range hub.connects {
		ret = append(ret, conn.id)
	}
	hub.mutex.RUnlock()

	return ret
}
