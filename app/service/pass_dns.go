package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
)

func NewPassDNS(qry *query.Query) *PassDNS {
	return &PassDNS{
		qry: qry,
	}
}

type PassDNS struct {
	qry *query.Query
}

func (biz *PassDNS) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.PassDNS) {
	tbl := biz.qry.PassDNS
	db := tbl.WithContext(ctx).UnderlyingDB().Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var ret []*model.PassDNS
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
