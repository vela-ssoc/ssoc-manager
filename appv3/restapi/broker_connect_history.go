package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type BrokerConnectHistory struct {
	svc *service.BrokerConnectHistory
}

func NewBrokerConnectHistory(svc *service.BrokerConnectHistory) *BrokerConnectHistory {
	return &BrokerConnectHistory{svc: svc}
}

func (bch *BrokerConnectHistory) RegisterHTTP(g shipx.RouteGroup) error {
	g.APIV3.Route("/broker/connect-histories").GET(bch.page)

	return nil
}

func (bch *BrokerConnectHistory) page(c *ship.Context) error {
	var param request.BrokerConnectHistoryPage
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	slug := param.Slug.Slug
	ctx := c.Request().Context()
	ret, err := bch.svc.Page(ctx, slug, param.PageSize)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
