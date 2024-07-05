package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
)

type CmdbService interface {
	Detail(ctx context.Context, id int64) *model.Cmdb2
}

func Cmdb() CmdbService {
	return &cmdbService{}
}

type cmdbService struct{}

func (biz *cmdbService) Detail(ctx context.Context, id int64) *model.Cmdb2 {
	tbl := query.Minion
	mon, err := tbl.WithContext(ctx).
		Select(tbl.Inet).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return nil
	}
	inet := mon.Inet

	cmdb2Tbl := query.Cmdb2
	dat, _ := cmdb2Tbl.WithContext(ctx).Where(cmdb2Tbl.Inet.Eq(inet)).First()

	return dat
}
