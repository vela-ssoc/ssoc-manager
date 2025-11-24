package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
)

func NewSBOMProject(qry *query.Query) *SBOMProject {
	return &SBOMProject{
		qry: qry,
	}
}

type SBOMProject struct{ qry *query.Query }

func (biz *SBOMProject) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMProject) {
	tbl := biz.qry.SBOMProject
	db := tbl.WithContext(ctx).
		Order(tbl.TotalScore.Desc()).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.SBOMProject
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}
