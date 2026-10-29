package restapi

import (
	"net/http"
	"net/http/httputil"
	"path"

	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/ssoc-common-mb/problem"
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
		ErrorHandler: problem.HandleError,
	}

	return &AgentReverseProxy{rt: rt, pxy: pxy}
}

func (arp *AgentReverseProxy) RegisterHTTP(g shipx.RouteGroup3) error {
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
	w, r := shipx.Strip(c, "path")
	r.URL.Path = path.Join("/api/v3/arp", r.URL.Path)
	r.URL.Scheme = "http"
	r.URL.Host = muxproto.BuildAgentFQDN(id)
	arp.pxy.ServeHTTP(w, r)

	return nil
}
