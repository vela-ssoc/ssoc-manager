package integrate

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/problem"
)

type ElasticSearcher interface {
	Load(ctx context.Context) (http.Handler, error)
	Reset()
}

func NewElastic(name string) ElasticSearcher {
	dialer := &net.Dialer{Timeout: 3 * time.Second}
	tlsDialer := &tls.Dialer{NetDialer: dialer}

	trip := &http.Transport{
		DialContext:    dialer.DialContext,
		DialTLSContext: tlsDialer.DialContext,
	}

	return &elasticProxy{
		name: name,
		trip: trip,
	}
}

type elasticProxy struct {
	name  string
	trip  http.RoundTripper
	mutex sync.RWMutex
	done  bool
	err   error
	px    *httputil.ReverseProxy
}

func (ela *elasticProxy) Load(ctx context.Context) (http.Handler, error) {
	ela.mutex.RLock()
	done, err, px := ela.done, ela.err, ela.px
	ela.mutex.RUnlock()
	if done {
		return px, err
	}
	return ela.loadSlow(ctx)
}

func (ela *elasticProxy) Reset() {
	ela.mutex.Lock()
	defer ela.mutex.Unlock()
	ela.err = nil
	ela.px = nil
	ela.done = false
}

func (ela *elasticProxy) loadSlow(ctx context.Context) (*httputil.ReverseProxy, error) {
	ela.mutex.Lock()
	defer ela.mutex.Unlock()

	if ela.done {
		return ela.px, ela.err
	}

	ela.done = true

	tbl := query.Elastic
	dat, err := tbl.WithContext(ctx).Where(tbl.Enable.Is(true)).First()
	if err != nil {
		ela.err = err
		return nil, err
	}

	px, err := ela.initProxy(dat.Host, dat.Username, dat.Password)
	ela.px = px
	ela.err = err

	return px, err
}

// initProxy 初始化创建代理，支持 BasicAuth
func (ela *elasticProxy) initProxy(addr, uname, passwd string) (*httputil.ReverseProxy, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	rewriteFunc := func(r *httputil.ProxyRequest) {
		r.SetXForwarded()
		r.SetURL(u)
		r.Out.SetBasicAuth(uname, passwd)
	}

	px := &httputil.ReverseProxy{
		Transport:      ela.trip,
		Rewrite:        rewriteFunc,
		ModifyResponse: ela.modifyResponse,
		ErrorHandler:   ela.errorHandler,
	}

	return px, nil
}

func (ela *elasticProxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	if e, ok := err.(*net.OpError); ok {
		// 隐藏后端服务 IP
		e.Addr = nil
		e.Net += " elasticsearch"
		err = e
	}

	pd := &problem.Detail{
		Type:     ela.name,
		Title:    "代理错误",
		Status:   http.StatusBadRequest,
		Detail:   err.Error(),
		Instance: r.RequestURI,
	}
	_ = pd.JSON(w)
}

func (ela *elasticProxy) modifyResponse(w *http.Response) error {
	if w.StatusCode == http.StatusUnauthorized {
		w.StatusCode = http.StatusBadRequest
	}
	return nil
}
