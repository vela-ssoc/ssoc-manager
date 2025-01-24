package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/applet/brkmux"
)

func NewPeer(qry *query.Query, log *slog.Logger) *Peer {
	return &Peer{
		qry: qry,
		log: log,
	}
}

type Peer struct {
	qry *query.Query
	log *slog.Logger
}

func (pee *Peer) Heartbeat(ctx context.Context, ident brkmux.Ident) error {
	now := time.Now()
	tbl := pee.qry.Broker
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(ident.ID)).
		UpdateSimple(tbl.HeartbeatAt.Value(now))

	return err
}
