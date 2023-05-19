package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/xgfone/ship/v5"
)

func Into(svc service.IntoService, headerKey, queryKey string) route.Router {
	return &intoREST{
		svc:       svc,
		headerKey: headerKey,
		queryKey:  queryKey,
	}
}

type intoREST struct {
	svc       service.IntoService
	headerKey string
	queryKey  string
}

func (ito *intoREST) Route(_, _, basic *ship.RouteGroupBuilder) {
	methods := []string{
		http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace,
		ship.PROPFIND, "LOCK", "MKCOL", "PROPPATCH", "COPY", "MOVE", "UNLOCK",
	}

	basic.Route("/bws/*path").GET(ito.BWS)
	basic.Route("/brr/*path").Method(ito.BRR, methods...)
	basic.Route("/aws/*path").GET(ito.AWS)
	basic.Route("/arr/*path").Method(ito.ARR, methods...)
}

func (ito *intoREST) BRR(c *ship.Context) error {
	if c.IsWebSocket() {
		return errcode.ErrUnsupportedWebSocket
	}

	node := ito.lookupNode(c)
	if node == "" {
		return errcode.ErrRequiredNode
	}

	w, r := c.Response(), c.Request()
	ctx := r.Context()
	ito.desensitization(r)

	return ito.svc.BRR(ctx, w, r, node)
}

func (ito *intoREST) BWS(c *ship.Context) error {
	if !c.IsWebSocket() {
		return errcode.ErrRequiredWebSocket
	}

	node := ito.lookupNode(c)
	if node == "" {
		return errcode.ErrRequiredNode
	}

	w, r := c.Response(), c.Request()
	ctx := r.Context()

	return ito.svc.BWS(ctx, w, r, node)
}

func (ito *intoREST) ARR(c *ship.Context) error {
	if c.IsWebSocket() {
		return errcode.ErrUnsupportedWebSocket
	}

	node := ito.lookupNode(c)
	if node == "" {
		return errcode.ErrRequiredNode
	}

	w, r := c.Response(), c.Request()
	ctx := r.Context()
	ito.desensitization(r)

	return ito.svc.ARR(ctx, w, r, node)
}

func (ito *intoREST) AWS(c *ship.Context) error {
	if !c.IsWebSocket() {
		return errcode.ErrRequiredWebSocket
	}

	node := ito.lookupNode(c)
	if node == "" {
		return errcode.ErrRequiredNode
	}

	w, r := c.Response(), c.Request()
	ctx := r.Context()

	return ito.svc.AWS(ctx, w, r, node)
}

func (ito *intoREST) lookupNode(c *ship.Context) (node string) {
	// Header > Query > Basic
	if node = c.GetReqHeader(linkhub.HeaderXNodeIdentify); node == "" {
		if node = c.Query(linkhub.QueryNodeKey); node == "" {
			node, _, _ = c.Request().BasicAuth()
		}
	}

	return node
}

// desensitization 对代理转发的请求脱敏，前端用户请求携带的认证信息不应该带到后面的节点请求中。
func (ito *intoREST) desensitization(r *http.Request) {
	r.Header.Del(ito.headerKey)
	if ito.headerKey != ship.HeaderAuthorization { // 删除 BasicAuth 信息
		r.Header.Del(ship.HeaderAuthorization)
	}

	// 删除 query 参数中的信息
	query := r.URL.Query()
	query.Del(ito.queryKey)
	r.URL.RawQuery = query.Encode()
}
