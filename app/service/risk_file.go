package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type RiskFileService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, model.RiskFiles)
}

func RiskFile() RiskFileService {
	return &riskFileService{}
}

type riskFileService struct{}

func (biz riskFileService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, model.RiskFiles) {
	tbl := query.RiskFile
	db := tbl.WithContext(ctx).UnderlyingDB().Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var ret model.RiskFiles
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
