package brkapi

import (
	"log/slog"

	"github.com/vela-ssoc/ssoc-common-mb/integration/dong/v2"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
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
	req := new(mrequest.BlinkAlert)
	if err := c.Bind(req); err != nil {
		return err
	}
	ctx := c.Request().Context()

	attrs := []any{"job_numbers", req.UserIDs, "group_ids", req.GroupIDs, "title", req.Title}
	alt.log.Info("准备发送咚咚告警", attrs...)
	err := alt.cli.Send(ctx, req.UserIDs, req.GroupIDs, req.Title, req.Detail)
	if err != nil {
		attrs = append(attrs, "err", err)
		alt.log.Error("告警发送失败", attrs...)
	} else {
		alt.log.Info("告警发送成功", attrs...)
	}

	return err
}
