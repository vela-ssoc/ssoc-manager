package service

import (
	"context"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-proto/muxconn"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type BrokerConnectNotice struct {
	fails *metrics.Counter
	succs *metrics.Counter
	discs *metrics.Counter
}

func NewBrokerConnectNotice() *BrokerConnectNotice {
	return &BrokerConnectNotice{
		fails: metrics.NewCounter("broker_tunnel_connect_fails"),
		succs: metrics.NewCounter("broker_tunnel_connect_succeeds"),
		discs: metrics.NewCounter("broker_tunnel_disconnects"),
	}
}

func (bcn *BrokerConnectNotice) OnFailed(ctx context.Context, mux muxconn.Muxer, err error) error {
	bcn.fails.Inc()
	return nil
}

func (bcn *BrokerConnectNotice) OnConnected(ctx context.Context, id bson.ObjectID, info muxserver.PeerInfo) error {
	bcn.succs.Inc()
	return nil
}

func (bcn *BrokerConnectNotice) OnDisconnected(ctx context.Context, id bson.ObjectID, info muxserver.PeerInfo, disconnectAt time.Time) error {
	bcn.discs.Inc()
	return nil
}
