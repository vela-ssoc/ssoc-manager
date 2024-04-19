package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type MinionCustomizedService interface {
	List(ctx context.Context) []*model.MinionCustomized
	Create(ctx context.Context, req *param.MinionCustomizedCreate) error
	Delete(ctx context.Context, id int64) error
}

func MinionCustomized() MinionCustomizedService {
	return &minionCustomizedService{}
}

type minionCustomizedService struct{}

func (svc *minionCustomizedService) List(ctx context.Context) []*model.MinionCustomized {
	tbl := query.MinionCustomized
	ret, err := tbl.WithContext(ctx).Order(tbl.ID).Find()
	if err != nil || ret == nil {
		return make([]*model.MinionCustomized, 0)
	}

	return ret
}

func (svc *minionCustomizedService) Create(ctx context.Context, req *param.MinionCustomizedCreate) error {
	// 查询定制总数
	tbl := query.MinionCustomized
	if count, _ := tbl.WithContext(ctx).Count(); count >= 100 {
		return errcode.ErrTooManyCustomized
	}
	// 查询标签名是否存在
	name := req.Name
	if count, _ := tbl.WithContext(ctx).Where(tbl.Name.Eq(name)).Count(); count != 0 {
		return errcode.FmtErrNameExist.Fmt(name)
	}

	// 保存数据
	dat := &model.MinionCustomized{Name: name, Icon: req.Icon}

	return tbl.WithContext(ctx).Create(dat)
}

func (svc *minionCustomizedService) Delete(ctx context.Context, id int64) error {
	tbl := query.MinionCustomized
	dat, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}

	binTbl := query.MinionBin
	if count, _ := binTbl.WithContext(ctx).
		Where(binTbl.Customized.Eq(dat.Name)).
		Count(); count != 0 {
		return errcode.ErrCustomizedUsed
	}

	_, err = tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Delete()

	return err
}
