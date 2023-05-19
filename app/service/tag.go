package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type TagService interface {
	Indices(ctx context.Context, idx param.Indexer) []string
}

func Tag() TagService {
	return &tagService{}
}

type tagService struct{}

func (biz *tagService) Indices(ctx context.Context, idx param.Indexer) []string {
	tbl := query.MinionTag
	dao := tbl.WithContext(ctx).Distinct(tbl.Tag)
	if kw := idx.Keyword(); kw != "" {
		dao.Where(tbl.Tag.Like(kw))
	}

	var dats []string
	_ = dao.Order(tbl.Tag).Scopes(idx.Scope).Scan(&dats)

	return dats
}
