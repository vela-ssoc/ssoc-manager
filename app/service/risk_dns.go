package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
)

func NewRiskDNS(qry *query.Query) *RiskDNS {
	return &RiskDNS{
		qry: qry,
	}
}

type RiskDNS struct {
	qry *query.Query
}

func (biz *RiskDNS) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.RiskDNS) {
	tbl := biz.qry.RiskDNS
	db := tbl.WithContext(ctx).UnderlyingDB().Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var ret []*model.RiskDNS
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
