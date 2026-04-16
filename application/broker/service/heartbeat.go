package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
)

type Heartbeat struct {
	qry *query.Query
	log *slog.Logger
}

func NewHeartbeat(qry *query.Query, log *slog.Logger) *Heartbeat {
	return &Heartbeat{
		qry: qry,
		log: log,
	}
}

func (hb *Heartbeat) Alive(ctx context.Context, id int64) error {
	tbl := hb.qry.Broker
	dao := tbl.WithContext(ctx)

	now := time.Now()
	_, err := dao.Where(tbl.ID.Eq(id), tbl.Status.Is(true)).
		UpdateSimple(tbl.HeartbeatAt.Value(now))

	return err
}
