package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type EventService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Event)
}

func Event() EventService {
	return &eventService{}
}

type eventService struct{}

func (evt *eventService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Event) {
	db := query.Event.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.Event
	db.Order("id DESC").Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}
