package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type DomainService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Domain)
}

func Domain() DomainService {
	return &domainService{}
}

type domainService struct{}

func (biz *domainService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Domain) {
	tbl := query.Domain
	db := tbl.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.Domain
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}
