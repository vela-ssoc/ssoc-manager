package muxaccept

import (
	"context"
	"time"

	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-proto/muxconn"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Notifier interface {
	OnFailed(ctx context.Context, mux muxconn.Muxer, err error) error
	OnConnected(ctx context.Context, id bson.ObjectID, info muxserver.PeerInfo) error
	OnDisconnected(ctx context.Context, id bson.ObjectID, info muxserver.PeerInfo, disconnectAt time.Time) error
}

func wrapSafeNotifier(ntf Notifier) *safeNotifier {
	return &safeNotifier{ntf: ntf}
}

type safeNotifier struct {
	ntf Notifier
}

func (s *safeNotifier) OnFailed(ctx context.Context, mux muxconn.Muxer, err error) error {
	if n := s.ntf; n != nil {
		return n.OnFailed(ctx, mux, err)
	}

	return nil
}

func (s *safeNotifier) OnConnected(ctx context.Context, id bson.ObjectID, info muxserver.PeerInfo) error {
	if n := s.ntf; n != nil {
		return n.OnConnected(ctx, id, info)
	}

	return nil
}

func (s *safeNotifier) OnDisconnected(ctx context.Context, id bson.ObjectID, info muxserver.PeerInfo, disconnectAt time.Time) error {
	if n := s.ntf; n != nil {
		return n.OnDisconnected(ctx, id, info, disconnectAt)
	}

	return nil
}
