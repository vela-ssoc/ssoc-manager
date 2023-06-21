package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type EventService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Event)
	Confirm(ctx context.Context, id []int64) error
	Delete(ctx context.Context, scope dynsql.Scope) error
}

func Event() EventService {
	return &eventService{}
}

type eventService struct{}

func (biz *eventService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.Event) {
	tbl := query.Event
	db := tbl.WithContext(ctx).
		Order(tbl.ID.Desc()).
		UnderlyingDB().
		Scopes(scope.Where)

	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.Event
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}

func (biz *eventService) Confirm(ctx context.Context, id []int64) error {
	tbl := query.Event
	ret, err := tbl.WithContext(ctx).
		Where(tbl.ID.In(id...), tbl.HaveRead.Is(false)).
		UpdateColumnSimple(tbl.HaveRead.Value(true))
	if err != nil || ret.RowsAffected != 0 {
		return err
	}

	return errcode.ErrOperateFailed
}

func (biz *eventService) Delete(ctx context.Context, scope dynsql.Scope) error {
	ret := query.Event.WithContext(ctx).
		UnderlyingDB().
		Scopes(scope.Where).
		Delete(&model.Event{})
	if ret.Error != nil || ret.RowsAffected != 0 {
		return ret.Error
	}

	return errcode.ErrDeleteFailed
}
