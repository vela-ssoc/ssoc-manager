package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type PassIPService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.PassIP)
}

func PassIP(qry *query.Query) PassIPService {
	return &passIPService{qry: qry}
}

type passIPService struct {
	qry *query.Query
}

func (biz *passIPService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.PassIP) {
	tbl := biz.qry.PassIP
	db := tbl.WithContext(ctx).UnderlyingDB().Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var ret []*model.PassIP
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
