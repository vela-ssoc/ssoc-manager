package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type RiskDNSService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.RiskDNS)
}

func RiskDNS(qry *query.Query) RiskDNSService {
	return &riskDNSService{
		qry: qry,
	}
}

type riskDNSService struct {
	qry *query.Query
}

func (biz *riskDNSService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.RiskDNS) {
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
