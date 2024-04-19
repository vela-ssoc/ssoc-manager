package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/query"
	"github.com/vela-ssoc/vela-common-mb-itai/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type ProcessService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionProcess)
}

func Process() ProcessService {
	return &processService{}
}

type processService struct{}

func (biz *processService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionProcess) {
	tbl := query.MinionProcess
	db := tbl.WithContext(ctx).
		Order(tbl.ID.Desc()).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if err := db.Count(&count).Error; err != nil || count == 0 {
		return 0, nil
	}

	ret := make([]*model.MinionProcess, 0, page.Size())
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
