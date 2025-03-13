package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
)

type EmailService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.Email)
	Create(ctx context.Context, req *mrequest.EmailCreate) error
	Update(ctx context.Context, req *mrequest.EmailUpdate) error
	Delete(ctx context.Context, id int64) error
}

func Email(qry *query.Query, pusher push.Pusher) EmailService {
	return &emailService{
		qry:    qry,
		pusher: pusher,
	}
}

type emailService struct {
	qry    *query.Query
	pusher push.Pusher
}

func (biz *emailService) Page(ctx context.Context, page param.Pager) (int64, []*model.Email) {
	tbl := biz.qry.Email
	dao := tbl.WithContext(ctx).
		Order(tbl.Enable.Desc(), tbl.ID)
	if kw := page.Keyword(); kw != "" {
		dao.Where(tbl.Host.Like(kw)).
			Or(tbl.Username.Like(kw))
	}
	count, _ := dao.Count()
	if count == 0 {
		return 0, nil
	}

	dats, _ := dao.Scopes(page.Scope(count)).Find()

	return count, dats
}

func (biz *emailService) Create(ctx context.Context, req *mrequest.EmailCreate) error {
	dat := &model.Email{
		Host:     req.Host,
		Username: req.Username,
		Password: req.Password,
		Enable:   req.Enable,
	}
	tbl := biz.qry.Email
	if err := tbl.WithContext(ctx).Create(dat); err != nil {
		return err
	}
	if req.Enable {
		biz.pusher.EmailReset(ctx)
	}

	return nil
}

func (biz *emailService) Update(ctx context.Context, req *mrequest.EmailUpdate) error {
	// 查询数据
	id := req.ID
	tbl := biz.qry.Email
	dat, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return err
	}

	reset := req.Enable || dat.Enable
	dat.Host = req.Host
	dat.Username = req.Username
	dat.Password = req.Password
	dat.Enable = req.Enable
	if !req.Enable {
		err = tbl.WithContext(ctx).Save(dat)
	} else {
		err = biz.qry.Transaction(func(tx *query.Query) error {
			if _, exx := tx.Email.WithContext(ctx).
				Where(tbl.Enable.Is(true)).
				UpdateSimple(tbl.Enable.Value(false)); exx != nil {
				return exx
			}
			return tx.Email.WithContext(ctx).Save(dat)
		})
	}

	if err == nil && reset {
		biz.pusher.EmailReset(ctx)
	}

	return err
}

func (biz *emailService) Delete(ctx context.Context, id int64) error {
	// 查询数据
	tbl := biz.qry.Email
	dat, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return err
	}
	// 删除数据
	if _, err = tbl.WithContext(ctx).
		Where(tbl.ID.Eq(id)).
		Delete(); err != nil {
		return err
	}

	if dat.Enable {
		biz.pusher.EmailReset(ctx)
	}

	return nil
}
