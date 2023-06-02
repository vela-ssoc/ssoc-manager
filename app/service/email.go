package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type EmailService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*model.Email)
	Create(ctx context.Context) error
	Update(ctx context.Context) error
	Delete(ctx context.Context) error
}

func Email() EmailService {
	return &emailService{}
}

type emailService struct{}

func (biz *emailService) Page(ctx context.Context, page param.Pager) (int64, []*model.Email) {
	tbl := query.Email
	dao := tbl.WithContext(ctx).Order(tbl.Enable.Desc())
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

func (biz *emailService) Create(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

func (biz *emailService) Update(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

func (biz *emailService) Delete(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}
