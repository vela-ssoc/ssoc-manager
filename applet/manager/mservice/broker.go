package mservice

import (
	"context"
	"log/slog"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/applet/brkmux"
)

func NewBroker(qry *query.Query, link brkmux.Linker, log *slog.Logger) *Broker {
	return &Broker{
		qry:  qry,
		link: link,
		log:  log,
	}
}

type Broker struct {
	qry  *query.Query
	link brkmux.Linker
	log  *slog.Logger
}

func (bk *Broker) Onlines() []brkmux.Broker {
	return bk.link.Brokers()
}

func (bk *Broker) Kickout(id int64) error {
	for _, brk := range bk.link.Brokers() {
		ident, _ := brk.Info()
		if ident.ID != id {
			continue
		}
		return brk.Close()
	}

	return nil
}

func (bk *Broker) ResetDB(ctx context.Context) error {
	tbl := bk.qry.Broker
	_, err := tbl.WithContext(ctx).
		Where(tbl.Status.Is(true)).
		UpdateSimple(tbl.Status.Value(false))

	return err
}
