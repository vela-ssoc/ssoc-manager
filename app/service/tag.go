package service

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type TagService interface {
	Indices(ctx context.Context, idx param.Indexer) []string
	Update(ctx context.Context, id int64, tags []string) error
}

func Tag(pusher push.Pusher) TagService {
	return &tagService{
		pusher: pusher,
	}
}

type tagService struct {
	pusher push.Pusher
}

func (biz *tagService) Indices(ctx context.Context, idx param.Indexer) []string {
	tbl := query.MinionTag
	dao := tbl.WithContext(ctx).Distinct(tbl.Tag)
	if kw := idx.Keyword(); kw != "" {
		dao.Where(tbl.Tag.Like(kw))
	}

	dats := make([]string, 0, idx.Size())
	_ = dao.Order(tbl.Tag).Scopes(idx.Scope).Scan(&dats)

	return dats
}

func (biz *tagService) Update(ctx context.Context, id int64, tags []string) error {
	monTbl := query.Minion
	mon, err := monTbl.WithContext(ctx).
		Select(monTbl.Status, monTbl.BrokerID, monTbl.Inet).
		Where(monTbl.ID.Eq(id)).
		First()
	if err != nil {
		return err
	}
	if mon.Status == model.MSDelete {
		return errcode.ErrNodeStatus
	}

	tbl := query.MinionTag
	// 查询现有的 tags
	olds, err := tbl.WithContext(ctx).Where(tbl.MinionID.Eq(id)).Find()
	if err != nil {
		return err
	}
	news := model.MinionTags(olds).Manual(id, tags)
	err = query.Q.Transaction(func(tx *query.Query) error {
		table := tx.WithContext(ctx).MinionTag
		if _, exx := table.Where(tbl.MinionID.Eq(id)).
			Delete(); exx != nil {
			return exx
		}
		return table.Clauses(clause.OnConflict{DoNothing: true}).
			CreateInBatches(news, 100)
	})

	if err == nil {
		biz.pusher.TaskSync(ctx, mon.BrokerID, id, mon.Inet)
	}

	return err
}
