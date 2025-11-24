package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
)

func NewDomain(qry *query.Query) *Domain {
	return &Domain{
		qry: qry,
	}
}

type Domain struct {
	qry *query.Query
}

func (biz *Domain) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Domain) {
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
