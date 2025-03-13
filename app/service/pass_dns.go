package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
)

type PassDNSService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.PassDNS)
}

func PassDNS(qry *query.Query) PassDNSService {
	return &passDNSService{
		qry: qry,
	}
}

type passDNSService struct {
	qry *query.Query
}

func (biz *passDNSService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.PassDNS) {
	tbl := biz.qry.PassDNS
	db := tbl.WithContext(ctx).UnderlyingDB().Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var ret []*model.PassDNS
	db.Scopes(page.DBScope(count)).Find(&ret)

	return count, ret
}
