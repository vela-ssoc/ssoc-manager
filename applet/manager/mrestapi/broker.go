package mrestapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-manager/applet/manager/mrequest"
	"github.com/vela-ssoc/vela-manager/applet/manager/mresponse"
	"github.com/vela-ssoc/vela-manager/applet/manager/mservice"
	"github.com/xgfone/ship/v5"
)

func NewBroker(svc *mservice.Broker) *Broker {
	return &Broker{
		svc: svc,
	}
}

type Broker struct {
	svc *mservice.Broker
}

func (bk *Broker) Route(r *ship.RouteGroupBuilder) error {
	r.Route("/broker/onlines").GET(bk.onlines)
	r.Route("/broker/kickout").GET(bk.kickout)
	return nil
}

func (bk *Broker) onlines(c *ship.Context) error {
	brks := bk.svc.Onlines()
	dats := make([]*mresponse.BrokerOnline, 0, len(brks))
	for _, brk := range brks {
		ident, issue := brk.Info()
		dats = append(dats, &mresponse.BrokerOnline{
			ID:   ident.ID,
			Name: issue.Name,
			Goos: ident.Goos,
			Arch: ident.Arch,
		})
	}

	return c.JSON(http.StatusOK, dats)
}

func (bk *Broker) kickout(c *ship.Context) error {
	req := new(mrequest.Int64ID)
	if err := c.BindQuery(req); err != nil {
		return err
	}

	return bk.svc.Kickout(req.ID)
}
