package linkhub

import (
	"github.com/vela-ssoc/ssoc-common-mb/param/negotiate"
	"github.com/vela-ssoc/ssoc-common-mba/smux"
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
