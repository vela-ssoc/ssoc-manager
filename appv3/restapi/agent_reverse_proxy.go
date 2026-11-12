package restapi

import (
	"net/http"
	"net/http/httputil"

	"github.com/vela-ssoc/ssoc-common-mb/httpship"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/xgfone/ship/v5"
)

type AgentReverseProxy struct {
	rt  http.RoundTripper
	pxy *httputil.ReverseProxy
}

func NewAgentReverseProxy(rt http.RoundTripper) *AgentReverseProxy {
	pxy := &httputil.ReverseProxy{
		Transport: rt,
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetXForwarded()
		},
		ErrorHandler: httpship.ErrorHandle,
	}

	return &AgentReverseProxy{rt: rt, pxy: pxy}
}

func (arp *AgentReverseProxy) RegisterHTTP(g shipx.RouteGroup) error {
	methods := []string{
		http.MethodOptions, http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodDelete,
		"QUERY", "LOCK", "UNLOCK", "PROPFIND", "PROPPATCH", "MKCOL", "COPY", "MOVE",
	}

	g.ExposeV1.Group.Route("/arr").Method(arp.handle, methods...)
	g.ExposeV1.Group.Route("/arr/*path").Method(arp.handle, methods...)

	return nil
}

func (arp *AgentReverseProxy) handle(c *ship.Context) error {
	id := c.GetReqHeader(muxproto.HeaderXNodeIdentify)
	if id == "" {
		return errcode.ErrMissingAgentNodeIdentify
	}

	// 隧道内部走 http 协议
	fqdn := muxproto.BuildAgentFQDN(id)
	w, r := c.Response(), c.Request()
	// r.URL.Path =  r.URL.Path
	r.URL.Scheme = "http"
	r.URL.Host = fqdn
	r.Host = fqdn
	arp.pxy.ServeHTTP(w, r)

	return nil
}
