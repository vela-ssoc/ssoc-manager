package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type MinionLogonService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionLogon)
}

func MinionLogon() MinionLogonService {
	return &minionLogonService{}
}

type minionLogonService struct{}

func (biz *minionLogonService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionLogon) {
	db := query.MinionLogon.WithContext(ctx).UnderlyingDB().Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.MinionLogon
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}
