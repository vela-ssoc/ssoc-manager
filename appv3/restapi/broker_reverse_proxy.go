package restapi

import (
	"net/http"
	"net/http/httputil"

	"github.com/vela-ssoc/ssoc-common-mb/problem"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type BrokerReverseProxy struct {
	rt  http.RoundTripper
	pxy *httputil.ReverseProxy
}

func NewBrokerReverseProxy(rt http.RoundTripper) *BrokerReverseProxy {
	pxy := &httputil.ReverseProxy{
		Transport: rt,
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetXForwarded()
		},
		ErrorHandler: problem.HandleError,
	}

	return &BrokerReverseProxy{rt: rt, pxy: pxy}
}

func (brp *BrokerReverseProxy) RegisterHTTP(g shipx.RouteGroup) error {
	methods := []string{
		http.MethodOptions, http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete,
		"QUERY", "LOCK", "UNLOCK", "PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE",
	}

	g.APIV3.Route("/br/:slug").Method(brp.handle, methods...)
	g.APIV3.Route("/br/:slug/*path").Method(brp.handle, methods...)

	return nil
}

func (brp *BrokerReverseProxy) handle(c *ship.Context) error {
	w, r := shipx.Strip(c, "path")
	slug := c.Param("slug")
	r.Header.Del(ship.HeaderAuthorization)
	r.URL.Scheme = "http"
	r.URL.Host = slug + ".broker.internal"

	brp.pxy.ServeHTTP(w, r)

	return nil
}
