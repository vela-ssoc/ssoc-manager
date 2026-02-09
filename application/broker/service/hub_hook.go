package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-proto/muxconn"
	"github.com/vela-ssoc/ssoc-proto/muxproto"
)

type HubHook struct {
	log *slog.Logger
}

func NewHubHook(log *slog.Logger) *HubHook {
	return &HubHook{
		log: log,
	}
}

func (hk *HubHook) OnAuthFailed(ctx context.Context, mux muxconn.Muxer, connAt time.Time, err error) {
	if ae, ok := err.(*muxproto.AuthError); ok {
		hk.log.Info("节点上线失败", "code", ae.Code, "message", ae.Text)
	} else {
		hk.log.Warn("节点上线失败了")
	}
}

func (hk *HubHook) OnConnected(ctx context.Context, inf muxserver.PeerInfo, connAt time.Time) {
}

func (hk *HubHook) OnDisconnected(ctx context.Context, inf muxserver.PeerInfo, connectAt, disconnectAt time.Time) {
}
