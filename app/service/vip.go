package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type VIPService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, model.VIPMembers)
}

func VIP() VIPService {
	return &vipService{}
}

type vipService struct{}

func (biz *vipService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, model.VIPMembers) {
	tbl := query.VIP
	db := tbl.WithContext(ctx).
		Distinct(tbl.VirtualAddr).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var addrs []string
	db.Scopes(page.DBScope(count)).Scan(&addrs)
	var vips model.VIPs
	db.Where("virtual_addr IN (?)", addrs).Find(&vips)
	dats := vips.Mapping()

	return count, dats
}
