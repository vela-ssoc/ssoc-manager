package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"gorm.io/gen/field"
)

type EmcService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.Emc)
	Create(ctx context.Context, ec *mrequest.EmcCreate) error
	Update(ctx context.Context, ec *mrequest.EmcUpdate) error
	Delete(ctx context.Context, id int64) error
}

func Emc(qry *query.Query, pusher push.Pusher) EmcService {
	return &emcService{
		qry:    qry,
		pusher: pusher,
	}
}

type emcService struct {
	qry    *query.Query
	pusher push.Pusher
}

func (biz *emcService) Page(ctx context.Context, page param.Pager) (int64, []*model.Emc) {
	tbl := biz.qry.Emc
	dao := tbl.WithContext(ctx).
		Order(tbl.Enable.Desc(), tbl.ID)
	if kw := page.Keyword(); kw != "" {
		dao.Where(tbl.Host.Like(kw)).
			Order(tbl.Account.Like(kw))
	}
	count, err := dao.Count()
	if err != nil || count == 0 {
		return 0, nil
	}
	dats, _ := dao.Scopes(page.Scope(count)).Find()

	return count, dats
}

func (biz *emcService) Create(ctx context.Context, ec *mrequest.EmcCreate) error {
	dat := &model.Emc{
		Name: ec.Name, Host: ec.Host, Account: ec.Account,
		Token: ec.Token, Enable: ec.Enable,
	}
	tbl := biz.qry.Emc
	if !ec.Enable {
		return tbl.WithContext(ctx).Create(dat)
	}

	err := biz.qry.Transaction(func(tx *query.Query) error {
		if _, exx := tx.WithContext(ctx).Emc.
			Where(tbl.Enable.Is(true)).
			UpdateColumnSimple(tbl.Enable.Value(false)); exx != nil {
			return exx
		}
		return tx.WithContext(ctx).Emc.Create(dat)
	})
	if err != nil {
		return err
	}

	// 重置所有节点的 emc 配置
	biz.pusher.EmcReset(ctx)

	return err
}

func (biz *emcService) Update(ctx context.Context, ec *mrequest.EmcUpdate) error {
	id, enable := ec.ID, ec.Enable
	tbl := biz.qry.Emc
	old, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}

	assigns := []field.AssignExpr{
		tbl.Name.Value(ec.Name),
		tbl.Host.Value(ec.Host),
		tbl.Account.Value(ec.Account),
		tbl.Token.Value(ec.Token),
		tbl.Enable.Value(ec.Enable),
	}
	if !enable && old.Enable {
		_, err = tbl.WithContext(ctx).
			Where(tbl.ID.Eq(id)).
			UpdateColumnSimple(assigns...)
		return err
	}

	err = biz.qry.Transaction(func(tx *query.Query) error {
		if _, exx := tx.WithContext(ctx).Emc.
			Where(tbl.Enable.Is(true)).
			UpdateColumnSimple(tbl.Enable.Value(false)); exx != nil {
			return exx
		}
		_, exx := tx.WithContext(ctx).Emc.
			Where(tbl.ID.Eq(id)).
			UpdateColumnSimple(assigns...)
		return exx
	})
	if err != nil {
		return err
	}

	// 重置所有节点的 emc 配置
	biz.pusher.EmcReset(ctx)

	return err
}

func (biz *emcService) Delete(ctx context.Context, id int64) error {
	tbl := biz.qry.Emc
	dat, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if _, err = tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		Delete(); err != nil || !dat.Enable {
		return err
	}

	// 重置所有节点的 emc 配置
	biz.pusher.EmcReset(ctx)

	return nil
}
