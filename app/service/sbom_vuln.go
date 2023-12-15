package service

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type SBOMVulnService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMVuln)
	Project(ctx context.Context, page param.Pager, purl string) (int64, []*model.SBOMProject)
	Vulnerabilities(ctx context.Context, offsetID int64, size int) []*model.SBOMVuln
	Purl(ctx context.Context, req *param.ReportPurl) error
}

func SBOMVuln() SBOMVulnService {
	return &sbomVulnService{}
}

type sbomVulnService struct{}

func (biz *sbomVulnService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.SBOMVuln) {
	tbl := query.SBOMVuln
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

func (biz *sbomVulnService) Project(ctx context.Context, page param.Pager, purl string) (int64, []*model.SBOMProject) {
	comTbl := query.SBOMComponent
	tbl := query.SBOMProject
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

func (biz *sbomVulnService) Vulnerabilities(ctx context.Context, offsetID int64, size int) []*model.SBOMVuln {
	tbl := query.SBOMVuln
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

func (biz *sbomVulnService) Purl(ctx context.Context, req *param.ReportPurl) error {
	ps := req.Purl
	dats := make([]*model.Purl, 0, len(ps))
	for _, p := range ps {
		dats = append(dats, &model.Purl{ID: p})
	}
	tbl := query.Purl

	return tbl.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: false}).Save(dats...)
}
