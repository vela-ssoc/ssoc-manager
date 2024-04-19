package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/query"
)

type CmdbService interface {
	Detail(ctx context.Context, id int64) *model.Cmdb
}

func Cmdb() CmdbService {
	return &cmdbService{}
}

type cmdbService struct{}

func (biz *cmdbService) Detail(ctx context.Context, id int64) *model.Cmdb {
	tbl := query.Cmdb
	dat, _ := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	return dat
}
