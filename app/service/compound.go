package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type CompoundService interface {
	Indices(ctx context.Context, idx param.Indexer) []*param.IDName
}

func Compound() CompoundService {
	return &compoundService{}
}

type compoundService struct{}

func (biz *compoundService) Indices(ctx context.Context, idx param.Indexer) []*param.IDName {
	tbl := query.Compound
	dao := tbl.WithContext(ctx).Select(tbl.ID, tbl.Name)
	if kw := idx.Keyword(); kw != "" {
		dao.Where(tbl.Name.Like(kw)).
			Or(tbl.Desc.Like(kw))
	}

	var dats []*param.IDName
	_ = dao.Order(tbl.ID).Scopes(idx.Scope).Scan(&dats)

	return dats
}
