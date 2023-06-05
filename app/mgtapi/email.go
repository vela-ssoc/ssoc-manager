package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func Email(svc service.EmailService) route.Router {
	return &emailREST{
		svc: svc,
	}
}

type emailREST struct {
	svc service.EmailService
}

func (rest *emailREST) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/emails").Data(route.Ignore()).GET(rest.Page)
}

func (rest *emailREST) Page(c *ship.Context) error {
	var req param.Page
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	page := req.Pager()
	count, dats := rest.svc.Page(ctx, page)
	res := page.Result(count, dats)

	return c.JSON(http.StatusOK, res)
}
