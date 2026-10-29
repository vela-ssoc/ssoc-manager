package muxlink

import (
	"github.com/vela-ssoc/vela-tunnel/muxconn"
)

type Node interface {
	MUX() muxconn.Muxer // 通信隧道
	Host() string       // 主机位，例如：test
	FQDN() string       // 全路径，例如：test.broker.internal
	Data() any          // 附带信息
}

type virtualNode struct {
	mux  muxconn.Muxer
	host string
	fqdn string
	data any
}

func (n *virtualNode) MUX() muxconn.Muxer { return n.mux }
func (n *virtualNode) Host() string       { return n.host }
func (n *virtualNode) FQDN() string       { return n.fqdn }
func (n *virtualNode) Data() any          { return n.data }
