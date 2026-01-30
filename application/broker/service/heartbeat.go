package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common/datalayer/query"
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

// Ping 处理 broker 节点发来的心跳包。
func (hb *Heartbeat) Ping(ctx context.Context, id int64) error {
	now := time.Now()
	tbl := hb.qry.Broker
	dao := tbl.WithContext(ctx)

	_, err := dao.Where(tbl.ID.Eq(id)).
		UpdateSimple(tbl.HeartbeatAt.Value(now))

	return err
}
