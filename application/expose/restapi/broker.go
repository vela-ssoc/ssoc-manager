package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/application/expose/request"
	"github.com/vela-ssoc/ssoc-manager/application/expose/service"
	"github.com/xgfone/ship/v5"
)

type Broker struct {
	svc *service.Broker
}

func NewBroker(svc *service.Broker) *Broker {
	return &Broker{
		svc: svc,
	}
}

func (b *Broker) BindRoute(rgb *ship.RouteGroupBuilder) error {
	rgb.Route("/brokers").GET(b.page)
	rgb.Route("/broker").POST(b.create)

	return nil
}

func (b *Broker) page(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, err := b.svc.Page(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (b *Broker) create(c *ship.Context) error {
	req := new(request.BrokerCreate)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return b.svc.Create(ctx, req)
}
