package linkhub

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/problem"
	"github.com/vela-ssoc/vela-common-mb/taskpool"
	"github.com/vela-ssoc/vela-common-mba/netutil"
	"github.com/vela-ssoc/vela-common-mba/spdy"
	"github.com/vela-ssoc/vela-manager/bridge/blink"
)

type Huber interface {
	blink.Joiner

	RestDB() error

	Unicast(id int64, path string, req, resp any) error

	Multicast(bids []int64, path string, req any) <-chan *ErrorFuture

	Broadcast(path, req any) <-chan *ErrorFuture

	// Stream 向 broker 节点建立 websocket 连接
	Stream(ctx context.Context, bid int64, path string, header http.Header) (*websocket.Conn, *http.Response, error)

	// Forward 向 broker 节点转发 http 请求
	Forward(bid int64, w http.ResponseWriter, r *http.Request)
}

func New(handler http.Handler, pool taskpool.Executor) Huber {
	hub := &brokerHub{
		name:     "manager",
		handler:  handler,
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
	client   netutil.HTTPClient
	forward  netutil.Forwarder
	streamer netutil.Streamer
	pool     taskpool.Executor
	mutex    sync.RWMutex
	connects map[string]*spdyServerConn
	random   *rand.Rand
}

func (hub *brokerHub) Name() string {
	return hub.name
}

func (hub *brokerHub) Auth(ident blink.Ident) (blink.Issue, http.Header, error) {
	// TODO implement me
	panic("implement me")
}

func (hub *brokerHub) Join(tran net.Conn, ident blink.Ident, issue blink.Issue) error {
	conn := hub.newConn(tran, ident, issue)
	if !hub.putConn(conn) {
		return errors.New("")
	}
	sid := conn.sid
	defer hub.delConn(sid)

	tbl := query.Broker
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	_, _ = tbl.WithContext(ctx).Where(tbl.ID.Eq(ident.ID)).UpdateColumn(tbl.Status, true)
	cancel()
	defer func() {
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
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
	_ = srv.Serve(conn.muxer) // 此处会阻塞，一旦执行结束说明连接断开

	return nil
}

func (hub *brokerHub) RestDB() error {
	brk := query.Broker
	_, err := brk.WithContext(context.Background()).
		Where(brk.Status.Is(true)).
		UpdateColumn(brk.Status, false)
	return err
}

func (hub *brokerHub) Unicast(id int64, path string, req, resp any) error {
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

func (hub *brokerHub) Multicast(bids []int64, path string, req any) <-chan *ErrorFuture {
	size := len(bids)
	ret := make(chan *ErrorFuture, size)
	if size == 0 {
		close(ret)
	} else {
		go hub.multicast(bids, path, req, ret)
	}

	return ret
}

func (hub *brokerHub) Broadcast(path, req any) <-chan *ErrorFuture {
	// TODO implement me
	panic("implement me")
}

func (hub *brokerHub) Forward(bid int64, w http.ResponseWriter, r *http.Request) {
	host := strconv.FormatInt(bid, 10)
	r.URL.Host = host
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
func (hub *brokerHub) silentJSON(id int64, path string, req any) error {
	addr := hub.httpURL(id, path)
	return hub.client.SilentJSON(nil, http.MethodPost, addr, req, nil)
}

// sendJSON 发送 JSON 请求响应 JSON 数据
func (hub *brokerHub) sendJSON(id int64, path string, req, resp any) error {
	addr := hub.httpURL(id, path)
	return hub.client.JSON(nil, http.MethodPost, addr, req, resp, nil)
}

func (hub *brokerHub) newConn(tran net.Conn, ident blink.Ident, issue blink.Issue) *spdyServerConn {
	id := ident.ID
	sid := strconv.FormatInt(id, 10)
	muxer := spdy.Server(tran, spdy.WithEncrypt(issue.Passwd))

	return &spdyServerConn{
		id:    id,
		sid:   sid,
		muxer: muxer,
		ident: ident,
		issue: issue,
	}
}

func (hub *brokerHub) httpURL(id int64, path string) string {
	sid := strconv.FormatInt(id, 10)
	u := &url.URL{Scheme: "http", Host: sid, Path: path}
	return u.String()
}

func (hub *brokerHub) wsURL(id int64, path string) string {
	sid := strconv.FormatInt(id, 10)
	u := &url.URL{Scheme: "ws", Host: sid, Path: path}
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
		return conn.muxer.Dial()
	}

	return nil, errors.New("broker 节点离线")
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
