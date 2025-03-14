package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
)

type CmdbService interface {
	Detail(ctx context.Context, id int64) *model.Cmdb2
}

func Cmdb(qry *query.Query) CmdbService {
	return &cmdbService{
		qry: qry,
	}
}

type cmdbService struct {
	qry *query.Query
}

func (biz *cmdbService) Detail(ctx context.Context, id int64) *model.Cmdb2 {
	tbl := biz.qry.Minion
	mon, err := tbl.WithContext(ctx).
		Select(tbl.Inet).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return nil
	}
	inet := mon.Inet

	cmdb2Tbl := biz.qry.Cmdb2
	dat, _ := cmdb2Tbl.WithContext(ctx).Where(cmdb2Tbl.Inet.Eq(inet)).First()

	return dat
}
