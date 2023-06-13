package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type PassDNSService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.PassDNS)
}

func PassDNS() PassDNSService {
	return &passDNSService{}
}

type passDNSService struct{}

func (biz *passDNSService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.PassDNS) {
	tbl := query.PassDNS
	db := tbl.WithContext(ctx).UnderlyingDB().Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var ret []*model.PassDNS
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
