package brkmux

import (
	"context"
	"time"

	"github.com/vela-ssoc/vela-common-mba/smux"
)

type Broker interface {
	Info() (Ident, Issue)
	Close() error
}

type brokerConn struct {
	mux         *smux.Session
	ident       Ident
	issue       Issue
	connectedAt time.Time
}

func (bc *brokerConn) Info() (Ident, Issue) {
	return bc.ident, bc.issue
}

func (bc *brokerConn) Close() error {
	return bc.mux.Close()
}

var brokerConnKey = &contextKey{name: "broker-conn"}

type contextKey struct {
	name string
}

func FromContext(ctx context.Context) Broker {
	if ctx == nil {
		return nil
	}
	val, _ := ctx.Value(brokerConnKey).(Broker)

	return val
}
