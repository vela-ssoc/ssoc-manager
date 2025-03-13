package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/service"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/xgfone/ship/v5"
)

func Elastic(svc service.ElasticService, headerKey string, queryKey string) route.Router {
	return &elasticREST{
		svc:       svc,
		headerKey: headerKey,
		queryKey:  queryKey,
	}
}

type elasticREST struct {
	svc       service.ElasticService
	headerKey string
	queryKey  string
}

func (ela *elasticREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/elastics").Data(route.Ignore()).GET(ela.Page)
	bearer.Route("/ribana/*path").Data(route.Named("Ribana")).Any(ela.Forward)
	bearer.Route("/ribana").Data(route.Named("Ribana")).Any(ela.Forward)
	bearer.Route("/elastic").
		Data(route.Named("新增 es 服务器")).POST(ela.Create).
		Data(route.Named("修改 es 服务器")).PUT(ela.Update).
		Data(route.Named("删除 es 服务器")).DELETE(ela.Delete)
	bearer.Route("/elastic/detect").
		Data(route.Named("探测 es 集群节点")).POST(ela.Detect)
}

func (ela *elasticREST) Forward(c *ship.Context) error {
	path := "/" + c.Param("path")
	w, r := c.Response(), c.Request()
	ctx := r.Context()
	r.URL.Path = path

	ela.desensitization(r)

	return ela.svc.Forward(ctx, w, r)
}

func (ela *elasticREST) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	page := req.Pager()
	ctx := c.Request().Context()
	count, dats := ela.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (ela *elasticREST) Create(c *ship.Context) error {
	var req mrequest.ElasticCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return ela.svc.Create(ctx, &req)
}

func (ela *elasticREST) Update(c *ship.Context) error {
	var req mrequest.ElasticUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	err := ela.svc.Update(ctx, &req)

	return err
}

func (ela *elasticREST) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return ela.svc.Delete(ctx, req.ID)
}

func (ela *elasticREST) Detect(c *ship.Context) error {
	var req mrequest.ElasticDetect
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	res := ela.svc.Detect(ctx, req.Host, req.Username, req.Password)

	return c.JSON(http.StatusOK, res)
}

// desensitization 对代理转发的请求脱敏，前端用户请求携带的认证信息不应该带到后面的节点请求中。
func (ela *elasticREST) desensitization(r *http.Request) {
	r.Header.Del(ela.headerKey)
	query := r.URL.Query()
	query.Del(ela.queryKey)
	r.URL.RawQuery = query.Encode()
}
