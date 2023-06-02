package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type MinionBinaryService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.MinionBin)
	Deprecate(ctx context.Context, id int64) error
	Delete(ctx context.Context, id int64) error
}

func MinionBinary() MinionBinaryService {
	return &minionBinaryService{}
}

type minionBinaryService struct{}

func (biz *minionBinaryService) Page(ctx context.Context, page param.Pager) (int64, []*model.MinionBin) {
	tbl := query.MinionBin
	dao := tbl.WithContext(ctx)
	if kw := page.Keyword(); kw != "" {
		dao.Where(tbl.Name.Like(kw)).
			Or(tbl.Goos.Like(kw)).
			Or(tbl.Semver.Like(kw))
	}
	count, err := dao.Count()
	if err != nil || count == 0 {
		return 0, nil
	}

	dats, _ := dao.Order(tbl.Weight.Desc()).
		Order(tbl.UpdatedAt.Desc()).
		Scopes(page.Scope(count)).
		Find()

	return count, dats
}

func (biz *minionBinaryService) Deprecate(ctx context.Context, id int64) error {
	tbl := query.MinionBin
	bin, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if bin.Deprecated {
		return errcode.ErrDeprecated
	}
	if _, err = tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		Where(tbl.Deprecated.Is(false)).
		UpdateColumnSimple(tbl.Deprecated.Value(true)); err != nil {
		return err
	}

	return err
}

func (biz *minionBinaryService) Delete(ctx context.Context, id int64) error {
	tbl := query.MinionBin
	_, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Delete()
	return err
}
