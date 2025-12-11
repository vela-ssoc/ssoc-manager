package service

import (
	"context"
	"log/slog"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
)

type Agent struct {
	qry *query.Query
	log *slog.Logger
}

func NewAgent(qry *query.Query, log *slog.Logger) *Agent {
	return &Agent{
		qry: qry,
		log: log,
	}
}

func (agt *Agent) Get(ctx context.Context, id int64) (*model.Minion, error) {
	tbl := agt.qry.Minion
	dao := tbl.WithContext(ctx)

	return dao.Where(tbl.ID.Eq(id)).First()
}
