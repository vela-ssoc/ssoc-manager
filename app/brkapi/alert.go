package brkapi

import (
	"github.com/vela-ssoc/vela-common-mb/integration/dong/v2"
	"github.com/vela-ssoc/vela-common-mb/logback"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/xgfone/ship/v5"
)

func NewAlert(cli dong.Client, log logback.Logger) *Alert {
	return &Alert{cli: cli, log: log}
}

type Alert struct {
	cli dong.Client
	log logback.Logger
}

func (alt *Alert) Router(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/alert/dong").POST(alt.dong)
}

func (alt *Alert) dong(c *ship.Context) error {
	alt.log.Info("发送咚咚告警")
	req := new(param.BlinkAlert)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()
	err := alt.cli.Send(ctx, req.UserIDs, req.GroupIDs, req.Title, req.Detail)
	if err != nil {
		alt.log.Errorf("告警发送失败：%s", err)
	}

	return err
}
