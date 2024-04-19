package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/query"
	"github.com/vela-ssoc/vela-common-mb-itai/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type RiskDNSService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.RiskDNS)
}

func RiskDNS() RiskDNSService {
	return &riskDNSService{}
}

type riskDNSService struct{}

func (biz *riskDNSService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.RiskDNS) {
	tbl := query.RiskDNS
	db := tbl.WithContext(ctx).UnderlyingDB().Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var ret []*model.RiskDNS
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
