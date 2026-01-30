package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common/datalayer/query"
	"github.com/vela-ssoc/ssoc-manager/config"
	"github.com/vela-ssoc/ssoc-manager/muxtunnel/muxaccept"
)

type Broker struct {
	cfg config.Database
	qry *query.Query
	log *slog.Logger
}

func NewBroker(cfg config.Database, qry *query.Query, log *slog.Logger) *Broker {
	return &Broker{
		cfg: cfg,
		qry: qry,
		log: log,
	}
}

func (bs *Broker) Reset(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tbl := bs.qry.Broker
	dao := tbl.WithContext(ctx)

	_, err := dao.Where(tbl.Status.Is(true)).
		UpdateSimple(tbl.Status.Value(false))

	return err
}

func (bs *Broker) LoadBoot(context.Context) (*muxaccept.BrokConfig, error) {
	c := bs.cfg
	return &muxaccept.BrokConfig{
		DSN:         c.DSN,
		MaxOpenConn: c.MaxOpenConn,
		MaxIdleConn: c.MaxIdleConn,
		MaxLifeTime: time.Duration(c.MaxLifeTime),
		MaxIdleTime: time.Duration(c.MaxIdleTime),
	}, nil
}
