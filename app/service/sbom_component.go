package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
)

func NewSBOMComponent(qry *query.Query) *SBOMComponent {
	return &SBOMComponent{
		qry: qry,
	}
}

type SBOMComponent struct {
	qry *query.Query
}

func (biz *SBOMComponent) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMComponent) {
	tbl := biz.qry.SBOMComponent
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

func (biz *SBOMComponent) Project(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMProject) {
	tbl := biz.qry.SBOMComponent
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

func (biz *SBOMComponent) Count(ctx context.Context, page param.Pager) (int64, request.NameCounts) {
	ret := make(request.NameCounts, 0, 10)
	tbl := biz.qry.SBOMComponent
	count, _ := tbl.WithContext(ctx).Distinct(tbl.Name).Count()
	if count == 0 {
		return 0, ret
	}

	_ = tbl.WithContext(ctx).
		Select(tbl.Name, tbl.Name.Count().As("count")).
		Group(tbl.Name).
		Order(tbl.Name.Count().Desc()).
		Scopes(page.Scope(count)).
		Scan(&ret)

	return count, ret
}
