package mgtapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/vela-ssoc/vela-manager/app/route"
	"github.com/vela-ssoc/vela-manager/app/service"
	"github.com/xgfone/ship/v5"
)

func NewMinionFilter(svc *service.MinionFilter) *MinionFilter {
	return &MinionFilter{
		svc: svc,
	}
}

type MinionFilter struct {
	svc *service.MinionFilter
}

func (mf *MinionFilter) Route(_, bearer, _ *ship.RouteGroupBuilder) {
	bearer.Route("/minions").Data(route.Ignore()).GET(mf.page)
	bearer.Route("/minion/cond").Data(route.Ignore()).GET(mf.cond)
	// bearer.Route("/minion-filter/search").Data(route.Ignore()).GET(mf.Search)
}

func (mf *MinionFilter) cond(c *ship.Context) error {
	ret := mf.svc.Cond()
	return c.JSON(http.StatusOK, ret)
}

func (mf *MinionFilter) page(c *ship.Context) error {
	req := new(request.PageKeywordConditions)
	if err := c.BindQuery(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	ret, err := mf.svc.Page(ctx, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
