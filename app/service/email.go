package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/push"
)

type EmailService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.Email)
	Create(ctx context.Context, req *param.EmailCreate) error
	Update(ctx context.Context, req *param.EmailUpdate) error
	Delete(ctx context.Context, id int64) error
}

func Email(pusher push.Pusher) EmailService {
	return &emailService{
		pusher: pusher,
	}
}

type emailService struct {
	pusher push.Pusher
}

func (biz *emailService) Page(ctx context.Context, page param.Pager) (int64, []*model.Email) {
	tbl := query.Email
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

func (biz *emailService) Create(ctx context.Context, req *param.EmailCreate) error {
	dat := &model.Email{
		Host:     req.Host,
		Username: req.Username,
		Password: req.Password,
		Enable:   req.Enable,
	}
	tbl := query.Email
	if err := tbl.WithContext(ctx).Create(dat); err != nil {
		return err
	}
	if req.Enable {
		biz.pusher.EmailReset(ctx)
	}

	return nil
}

func (biz *emailService) Update(ctx context.Context, req *param.EmailUpdate) error {
	// 查询数据
	id := req.ID
	tbl := query.Email
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
		err = query.Q.Transaction(func(tx *query.Query) error {
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
	tbl := query.Email
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
