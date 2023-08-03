package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func BrokerBinary(svc service.BrokerBinaryService) route.Router {
	return &brokerBinaryREST{
		svc: svc,
	}
}

type brokerBinaryREST struct {
	svc service.BrokerBinaryService
}

func (rest *brokerBinaryREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/brkbins").Data(route.Ignore()).GET(rest.Page)
	bearer.Route("/brkbin").
		Data(route.Ignore()).POST(rest.Download).
		Data(route.Named("上传 broker 客户端")).POST(rest.Create).
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
	var req param.IntID
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
	var req param.NodeBinaryCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	return rest.svc.Create(ctx, &req)
}
