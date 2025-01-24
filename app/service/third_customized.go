package service

import (
	"context"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type ThirdCustomizedService interface {
	List(ctx context.Context) []*model.ThirdCustomized
	Create(ctx context.Context, req *param.ThirdCustomizedCreate) error
	Update(ctx context.Context, req *param.ThirdCustomizedUpdate) error
	Delete(ctx context.Context, id int64) error
}

func ThirdCustomized(qry *query.Query) ThirdCustomizedService {
	return &thirdCustomizedService{qry: qry}
}

type thirdCustomizedService struct {
	qry *query.Query
}

func (svc *thirdCustomizedService) List(ctx context.Context) []*model.ThirdCustomized {
	tbl := svc.qry.ThirdCustomized
	ret, err := tbl.WithContext(ctx).Order(tbl.ID).Find()
	if err != nil || ret == nil {
		return make([]*model.ThirdCustomized, 0)
	}
	return ret
}

func (svc *thirdCustomizedService) Create(ctx context.Context, req *param.ThirdCustomizedCreate) error {
	// 查询定制总数
	tbl := svc.qry.ThirdCustomized
	if count, _ := tbl.WithContext(ctx).Count(); count >= 100 {
		return errcode.ErrTooManyCustomized
	}
	// 查询标签名是否存在
	name := req.Name
	if count, _ := tbl.WithContext(ctx).Where(tbl.Name.Eq(name)).Count(); count != 0 {
		return errcode.FmtErrNameExist.Fmt(name)
	}

	// 保存数据
	dat := &model.ThirdCustomized{Name: name, Icon: req.Icon, Remark: req.Remark}

	return tbl.WithContext(ctx).Create(dat)
}

func (svc *thirdCustomizedService) Update(ctx context.Context, req *param.ThirdCustomizedUpdate) error {
	// 查询定制总数
	tbl := svc.qry.ThirdCustomized
	if count, _ := tbl.WithContext(ctx).Count(); count >= 100 {
		return errcode.ErrTooManyCustomized
	}
	_, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(req.ID)).
		UpdateSimple(
			tbl.UpdatedAt.Value(time.Now()),
			tbl.Icon.Value(req.Icon),
			tbl.Remark.Value(req.Remark),
		)

	return err
}

func (svc *thirdCustomizedService) Delete(ctx context.Context, id int64) error {
	tbl := svc.qry.ThirdCustomized
	dat, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}

	thrTbl := svc.qry.Third
	if count, _ := thrTbl.WithContext(ctx).
		Where(thrTbl.Customized.Eq(dat.Name)).
		Count(); count != 0 {
		return errcode.ErrCustomizedUsed
	}
	_, err = tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Delete()

	return err
}
