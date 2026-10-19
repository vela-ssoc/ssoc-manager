package muxcentral

import (
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/muxproto"
)

type sessionData struct {
	Request     muxproto.BrokerAuthRequest
	TunnelStat  model.TunnelStat
	ExecuteStat model.ExecuteStat
	ConnectedAt time.Time
}
