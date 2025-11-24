package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"gorm.io/gorm/clause"
)

func NewSBOMVuln(qry *query.Query) *SBOMVuln {
	return &SBOMVuln{qry: qry}
}

type SBOMVuln struct{ qry *query.Query }

func (biz *SBOMVuln) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMVuln) {
	tbl := biz.qry.SBOMVuln
	db := tbl.WithContext(ctx).
		Order(tbl.Score.Desc()).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.SBOMVuln
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}

func (biz *SBOMVuln) Project(ctx context.Context, page param.Pager, purl string) (int64, []*model.SBOMProject) {
	comTbl := biz.qry.SBOMComponent
	tbl := biz.qry.SBOMProject
	subSQL := comTbl.WithContext(ctx).
		Distinct(comTbl.ProjectID).
		Where(comTbl.PURL.Eq(purl))

	dao := tbl.WithContext(ctx).
		Where(tbl.WithContext(ctx).Columns(tbl.ID).In(subSQL))
	count, err := dao.Count()
	if err != nil || count == 0 {
		return 0, nil
	}
	dats, _ := dao.Order(tbl.TotalScore).
		Order(tbl.ID).
		Scopes(page.Scope(count)).
		Find()

	return count, dats
}

func (biz *SBOMVuln) Vulnerabilities(ctx context.Context, offsetID int64, size int) []*model.SBOMVuln {
	tbl := biz.qry.SBOMVuln
	ret, _ := tbl.WithContext(ctx).
		Where(tbl.ID.Gt(offsetID)).
		Limit(size).
		Order(tbl.ID).
		Find()
	if len(ret) == 0 {
		return []*model.SBOMVuln{}
	}

	return ret
}

func (biz *SBOMVuln) Purl(ctx context.Context, req *param.ReportPurl) error {
	ps := req.Purl
	dats := make([]*model.Purl, 0, len(ps))
	for _, p := range ps {
		dats = append(dats, &model.Purl{ID: p})
	}
	tbl := biz.qry.Purl

	return tbl.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: false}).Save(dats...)
}
