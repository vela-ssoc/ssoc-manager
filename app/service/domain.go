package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
)

type DomainService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Domain)
}

func Domain(qry *query.Query) DomainService {
	return &domainService{
		qry: qry,
	}
}

type domainService struct {
	qry *query.Query
}

func (biz *domainService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Domain) {
	tbl := biz.qry.Domain
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
