package service

import (
	"log/slog"

	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
	"github.com/vela-ssoc/vela-tunnel/muxconn"
)

type BrokerStream struct {
	hub muxproto.Huber
	log *slog.Logger
}

func NewBrokerStream(hub muxproto.Huber, log *slog.Logger) *BrokerStream {
	return &BrokerStream{hub: hub, log: log}
}

func (bs *BrokerStream) Substreams(slug string) ([]muxconn.Streamer, bool) {
	peer := bs.hub.Get(slug)
	if peer == nil {
		return nil, false
	}

	mux := peer.MUX()
	streams := mux.Streams()

	return streams, true
}

func (bs *BrokerStream) Peer(slug string) muxproto.Peer {
	return bs.hub.Get(slug)
}
