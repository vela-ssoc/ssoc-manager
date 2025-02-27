package linkhub

import (
	"github.com/vela-ssoc/vela-common-mb/param/negotiate"
	"github.com/vela-ssoc/vela-common-mba/smux"
)

type spdyServerConn struct {
	id    int64
	sid   string
	muxer *smux.Session
	ident negotiate.Ident
	issue negotiate.Issue
}

func (sc *spdyServerConn) ID() int64 {
	return sc.id
}
