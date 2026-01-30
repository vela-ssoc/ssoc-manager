package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/brkclient"
	"github.com/xgfone/ship/v5"
)

type Broker struct {
	cli brkclient.Client
}

func NewBroker(cli brkclient.Client) *Broker {
	return &Broker{cli: cli}
}

func (brk *Broker) BindRoute(rgb *ship.RouteGroupBuilder) error {
	rgb.Route("/broker/tunnel/stat").GET(brk.tunnelStat)
	rgb.Route("/broker/tunnel/limit").POST(brk.tunnelLimit)

	return nil
}

func (brk *Broker) tunnelStat(c *ship.Context) error {
	req := new(request.Int64ID)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, err := brk.cli.TunnelStat(ctx, req.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (brk *Broker) tunnelLimit(c *ship.Context) error {
	return nil
}
