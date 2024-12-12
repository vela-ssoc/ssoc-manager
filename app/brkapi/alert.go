package brkapi

import (
	"github.com/vela-ssoc/vela-common-mb/integration/dong"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/xgfone/ship/v5"
)

func NewAlert(cli dong.Client) *Alert {
	return &Alert{cli: cli}
}

type Alert struct {
	cli dong.Client
}

func (alt *Alert) Router(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/alert/dong").POST(alt.dong)
}

func (alt *Alert) dong(c *ship.Context) error {
	req := new(param.BlinkAlert)
	if err := c.Bind(req); err != nil {
		return err
	}

	return nil
}
