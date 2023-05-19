package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type SBOMService interface {
	PageComponent(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMComponent)
}

func SBOM() SBOMService {
	return &sbomService{}
}

type sbomService struct{}

func (biz *sbomService) PageComponent(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMComponent) {
	tbl := query.SBOMComponent
	db := tbl.WithContext(ctx).Order(tbl.TotalScore.Desc()).UnderlyingDB().Scopes(scope.Where)

	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.SBOMComponent
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}
