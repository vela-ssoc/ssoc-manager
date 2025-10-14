package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
)

func NewRiskFile(qry *query.Query) *RiskFile {
	return &RiskFile{qry: qry}
}

type RiskFile struct {
	qry *query.Query
}

func (biz RiskFile) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, model.RiskFiles) {
	tbl := biz.qry.RiskFile
	db := tbl.WithContext(ctx).UnderlyingDB().Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var ret model.RiskFiles
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
