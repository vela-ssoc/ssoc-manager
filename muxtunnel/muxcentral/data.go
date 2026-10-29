package muxcentral

import (
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
)

type sessionData struct {
	Model       *model.Broker
	Request     muxproto.BrokerAuthRequest
	TunnelStat  model.TunnelStat
	ExecuteStat model.ExecuteStat
	ConnectedAt time.Time
}

type EventData struct {
	Slug           string    `json:"slug"`
	Name           string    `json:"name"`
	Inet           string    `json:"inet"`
	Version        string    `json:"version"`
	Text           string    `json:"text,omitzero"`
	ConnectedAt    time.Time `json:"connected_at,omitzero"`
	DisconnectedAt time.Time `json:"disconnected_at,omitzero"`
}
