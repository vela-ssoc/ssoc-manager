package mgtapi

import (
	"net"
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/vela-ssoc/vela-manager/param/mrequest"
	"github.com/xgfone/ship/v5"
)

func BrokerBinary(svc *service.BrokerBinary) route.Router {
	return &brokerBinaryREST{
		svc: svc,
	}
}

type brokerBinaryREST struct {
	svc *service.BrokerBinary
}

func (rest *brokerBinaryREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/brkbins").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/brkbin/latest").Data(route.Ignore()).GET(rest.latest)
	bearer.Route("/brkbin/supports").Data(route.Ignore()).GET(rest.supports)
	bearer.Route("/brkbin").
		Data(route.Ignore()).GET(rest.Download).
		Data(route.IgnoreBody("上传 broker 客户端")).POST(rest.Create).
		Data(route.Named("删除 broker 客户端")).DELETE(rest.Delete)
}

func (rest *brokerBinaryREST) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}
	page := req.Pager()
	ctx := c.Request().Context()

	count, dats := rest.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}

func (rest *brokerBinaryREST) Delete(c *ship.Context) error {
	var req request.Int64ID
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Delete(ctx, req.ID)
}

func (rest *brokerBinaryREST) Create(c *ship.Context) error {
	var req param.NodeBinaryCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}

func (rest *brokerBinaryREST) Download(c *ship.Context) error {
	var req mrequest.BrokerDownload
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	// 获取内网地址
	ctx := c.Request().Context()

	var addr net.Addr
	val := ctx.Value(http.LocalAddrContextKey)
	if a, ok := val.(net.Addr); ok {
		addr = a
	}

	host := c.Host()
	if hostHeader, _, exx := net.SplitHostPort(host); exx == nil && hostHeader != "" {
		host = hostHeader
	}
	file, err := rest.svc.Open(ctx, req.BrokerID, req.ID, addr, host)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	// 此时的 Content-Length = 原始文件 + 隐藏文件
	c.Header().Set(ship.HeaderContentLength, file.ContentLength())
	c.Header().Set(ship.HeaderContentDisposition, file.Disposition())

	return c.Stream(http.StatusOK, file.ContentType(), file)
}

func (rest *brokerBinaryREST) latest(c *ship.Context) error {
	req := new(mrequest.BrokerBinaryLatest)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	bin := rest.svc.Latest(ctx, req.Goos, req.Arch)

	return c.JSON(http.StatusOK, bin)
}

func (rest *brokerBinaryREST) supports(c *ship.Context) error {
	dat := rest.svc.Supports()
	return c.JSON(http.StatusOK, dat)
}
