package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type SBOMComponentService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMComponent)
	Project(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMProject)
}

func SBOMComponent() SBOMComponentService {
	return &sbomComponentService{}
}

type sbomComponentService struct{}

func (biz *sbomComponentService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMComponent) {
	tbl := query.SBOMComponent
	db := tbl.WithContext(ctx).
		Order(tbl.TotalScore.Desc()).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.SBOMComponent
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}

func (biz *sbomComponentService) Project(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMProject) {
	tbl := query.SBOMComponent
	db := tbl.WithContext(ctx).UnderlyingDB()
	subSQL := db.Model(&model.SBOMComponent{}).
		Distinct("project_id").
		Scopes(scope.Where)

	tx := db.Model(&model.SBOMProject{}).Where("id IN (?)", subSQL)
	var count int64
	if tx.Count(&count); count == 0 {
		return 0, nil
	}
	var dats []*model.SBOMProject
	tx.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}
