package brkapi

import (
	"log/slog"

	"github.com/vela-ssoc/vela-common-mb/integration/dong/v2"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/xgfone/ship/v5"
)

func NewAlert(cli dong.Client, log *slog.Logger) *Alert {
	return &Alert{cli: cli, log: log}
}

type Alert struct {
	cli dong.Client
	log *slog.Logger
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
		alt.log.Error("告警发送失败", slog.Any("error", err))
	}

	return err
}
