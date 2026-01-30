package service

import (
	"context"
	"log/slog"

	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
)

type BrokerStatus struct {
	qry *query.Query
	log *slog.Logger
}

func NewBrokerStatus(qry *query.Query, log *slog.Logger) *BrokerStatus {
	return &BrokerStatus{
		qry: qry,
		log: log,
	}
}

func (bs *BrokerStatus) Reset(ctx context.Context) error {
	tbl := bs.qry.Broker
	dao := tbl.WithContext(ctx)

	_, err := dao.Where(tbl.Status.Is(true)).
		UpdateSimple(tbl.Status.Value(false))

	return err
}
