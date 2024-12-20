package brkapi

import (
	"context"
	"net/http"

	"github.com/xgfone/ship/v5"
)

type SIEMHandler interface {
	SIEMHandler(ctx context.Context) (http.Handler, error)
}

func NewSIEM(handler SIEMHandler) *SIEM {
	return &SIEM{handler: handler}
}

type SIEM struct {
	handler SIEMHandler
}

func (sim *SIEM) Router(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/proxy/siem").Any(sim.proxy)
	rgb.Route("/proxy/siem/*path").Any(sim.proxy)
}

func (sim *SIEM) proxy(c *ship.Context) error {
	path := "/" + c.Param("path")
	w, r := c.Response(), c.Request()
	r.URL.Path = path
	ctx := r.Context()
	handler, err := sim.handler.SIEMHandler(ctx)
	if err != nil {
		return err
	}
	handler.ServeHTTP(w, r)

	return nil
}
