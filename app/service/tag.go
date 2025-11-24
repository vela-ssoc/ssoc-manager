package service

import (
	"context"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/bridge/push"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"gorm.io/gen"
	"gorm.io/gorm/clause"
)

func NewTag(qry *query.Query, pusher push.Pusher) *Tag {
	return &Tag{
		qry:    qry,
		pusher: pusher,
	}
}

type Tag struct {
	qry    *query.Query
	pusher push.Pusher
}

func (biz *Tag) Indices(ctx context.Context, idx param.Indexer) []string {
	tbl := biz.qry.MinionTag

	dao := tbl.WithContext(ctx).
		Distinct(tbl.Tag).
		Order(tbl.Tag)
	if kw := idx.Keyword(); kw != "" {
		dao.Where(tbl.Tag.Like(kw))
	} else {
		left := int8(model.TkLifelong)
		dao.Not(tbl.Kind.Eq(left))
	}

	dats := make([]string, 0, 50)
	_ = dao.Order(tbl.Tag).Scan(&dats)

	return dats
}

func (biz *Tag) Update(ctx context.Context, id int64, tags []string) error {
	monTbl := biz.qry.Minion
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

	tbl := biz.qry.MinionTag
	// 查询现有的 tags
	olds, err := tbl.WithContext(ctx).Where(tbl.MinionID.Eq(id)).Find()
	if err != nil {
		return err
	}
	news := model.MinionTags(olds).Manual(id, tags)
	err = biz.qry.Transaction(func(tx *query.Query) error {
		table := tx.WithContext(ctx).MinionTag
		if _, exx := table.Where(tbl.MinionID.Eq(id)).
			Delete(); exx != nil {
			return exx
		}
		return table.Clauses(clause.OnConflict{DoNothing: true}).
			CreateInBatches(news, 100)
	})

	// 标签发生修改，则意味着关联的配置发生了修改
	if err == nil {
		biz.pusher.TaskSync(ctx, mon.BrokerID, []int64{id})
	}

	return err
}

func (biz *Tag) Sidebar(ctx context.Context, req *param.TagSidebar) (request.NameCounts, error) {
	tbl := biz.qry.MinionTag
	dao := tbl.WithContext(ctx)
	var conds []gen.Condition
	if kw := req.Keyword; kw != "" {
		kw = "%" + kw + "%"
		conds = append(conds, tbl.Tag.Like(kw))
	}
	//if !req.IPv4 {
	//	lifelong := int8(model.TkLifelong)
	//	regex := "^(?:[0-9]{1,3}\\.){3}[0-9]{1,3}$"
	//	or := field.Or(tbl.Kind.Eq(lifelong), tbl.Tag.NotRegxp(regex))
	//	conds = append(conds, tbl.Kind.Neq(lifelong), or)
	//}

	limit := 50
	ret := make(request.NameCounts, 0, limit)
	name, count := ret.Aliases()
	nameAlias := name.ColumnName().String()
	countAlias := count.ColumnName().String()

	if err := dao.Where(conds...).
		Select(tbl.Tag.As(nameAlias), tbl.Tag.Count().As(countAlias)).
		Group(tbl.Tag).
		Order(count.Desc()).
		Limit(limit).
		Scan(&ret); err != nil {
		return nil, err
	}

	return ret, nil
}
