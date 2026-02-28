package muxaccept

import (
	"time"

	"github.com/vela-ssoc/ssoc-common/muxserver"
	"github.com/vela-ssoc/ssoc-common/store/model"
	"github.com/vela-ssoc/ssoc-proto/muxproto"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type tunnelSessionData struct {
	ID            bson.ObjectID               `json:"id,omitzero"`
	Peer          muxserver.Peer              `json:"-"`
	Name          string                      `json:"name,omitzero"`
	Request       *muxproto.BrokerAuthRequest `json:"-"`
	ConnectAt     time.Time                   `json:"connect_at,omitzero"`
	DisconnectAt  time.Time                   `json:"disconnect_at,omitzero"`
	LocalAddr     string                      `json:"local_addr,omitzero"`
	RemoteAddr    string                      `json:"remote_addr,omitzero"`
	TunnelLibrary model.TunnelLibrary         `json:"tunnel_library,omitzero"`
	ExecuteStat   model.ExecuteStat           `json:"execute_stat,omitzero"`
	TunnelStat    model.TunnelStat            `json:"tunnel_stat,omitzero"`
}

func (d tunnelSessionData) connectedSeconds() uint64 {
	return uint64(d.DisconnectAt.Sub(d.ConnectAt).Seconds())
}
