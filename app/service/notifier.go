package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type NotifierService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.Notifier)
	Create(ctx context.Context, req *param.NotifierCreate) error
	Update(ctx context.Context, req *param.NotifierUpdate) error
	Delete(ctx context.Context, id int64) error
}

func Notifier(pusher push.Pusher) NotifierService {
	return &notifierService{
		pusher: pusher,
	}
}

type notifierService struct {
	pusher push.Pusher
}

func (biz *notifierService) Page(ctx context.Context, page param.Pager) (int64, []*model.Notifier) {
	tbl := query.Notifier
	dao := tbl.WithContext(ctx)
	if kw := page.Keyword(); kw != "" {
		dao = dao.Or(tbl.Name.Like(kw), tbl.Mobile.Like(kw), tbl.Email.Like(kw), tbl.Dong.Like(kw))
	}
	count, _ := dao.Count()
	if count == 0 {
		return 0, nil
	}

	dats, _ := dao.Scopes(page.Scope(count)).Find()

	return count, dats
}

func (biz *notifierService) Create(ctx context.Context, req *param.NotifierCreate) error {
	dat := &model.Notifier{
		Name:      req.Name,
		Events:    req.Events,
		Risks:     req.Risks,
		Ways:      req.Ways,
		Dong:      req.Dong,
		Email:     req.Email,
		Mobile:    req.Mobile,
		EventCode: req.EventCode,
		RiskCode:  req.RiskCode,
	}

	err := query.Notifier.WithContext(ctx).Create(dat)
	if err == nil {
		biz.pusher.NotifierReset(ctx)
	}

	return err
}

func (biz *notifierService) Update(ctx context.Context, req *param.NotifierUpdate) error {
	tbl := query.Notifier
	dat, err := tbl.WithContext(ctx).
		Where(tbl.ID.Eq(req.ID)).
		First()
	if err != nil {
		return err
	}

	dat.Name = req.Name
	dat.Events = req.Events
	dat.Risks = req.Risks
	dat.Ways = req.Ways
	dat.Dong = req.Dong
	dat.Email = req.Email
	dat.Mobile = req.Mobile
	dat.EventCode = req.EventCode
	dat.RiskCode = req.RiskCode

	if err = tbl.WithContext(ctx).
		Where(tbl.ID.Eq(req.ID)).
		Save(dat); err != nil {
		return err
	}
	biz.pusher.NotifierReset(ctx)

	return nil
}

func (biz *notifierService) Delete(ctx context.Context, id int64) error {
	tbl := query.Notifier
	ret, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Delete()
	if err != nil {
		return err
	}
	if ret.RowsAffected == 0 {
		return errcode.ErrDeleteFailed
	}
	biz.pusher.NotifierReset(ctx)

	return nil
}
