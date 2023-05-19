package linkhub

import (
	"github.com/vela-ssoc/vela-common-mba/spdy"
	"github.com/vela-ssoc/vela-manager/bridge/blink"
)

type spdyServerConn struct {
	id    int64
	sid   string
	muxer spdy.Muxer
	ident blink.Ident
	issue blink.Issue
}

func (sc *spdyServerConn) ID() int64 {
	return sc.id
}
