package service

import (
	"context"
	"log/slog"

	model2 "github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
)

type Minion struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewMinion(pdb polydb.Sourcer, log *slog.Logger) *Minion {
	return &Minion{pdb: pdb, log: log}
}

func (mon *Minion) Page(ctx context.Context, param request.MinionPage) (*repository.Pages[model2.Minion], error) {
	tbl := mon.pdb.RDB().Minion
	dao := tbl.WithContext(ctx)

	if dept := param.Department; dept != "" {
		dao.Where(tbl.OrgPath.Eq(dept)).Or(tbl.OrgPath.Like(dept + "/%"))
	}
	page, size := param.PSN()
	count, err := dao.Count()
	if err != nil {
		return nil, err
	} else if count == 0 {
		return repository.NewEmptyPages[model2.Minion](size), nil
	}
	records, err := dao.Scopes(param.Scope(count)).Find()
	if err != nil {
		return nil, err
	}
	ret := &repository.Pages[model2.Minion]{
		Page:    page,
		Size:    size,
		Total:   count,
		Records: records,
	}

	return ret, nil
}
