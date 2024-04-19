package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb-itai/dal/model"
	"github.com/vela-ssoc/vela-common-mb-itai/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
	"gorm.io/gorm/clause"
)

type TagService interface {
	Indices(ctx context.Context, idx param.Indexer) []string
	Update(ctx context.Context, id int64, tags []string) error
	Sidebar(ctx context.Context) []*param.NameCount
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

	// 标签发生修改，则意味着关联的配置发生了修改
	if err == nil {
		biz.pusher.TaskSync(ctx, mon.BrokerID, []int64{id})
	}

	return err
}

func (biz *tagService) Sidebar(ctx context.Context) []*param.NameCount {
	tbl := query.MinionTag
	// lifelong := int8(model.TkLifelong)
	ipv4 := `^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`
	ret := make([]*param.NameCount, 0, 100)
	// tbl.WithContext(ctx).
	//	Where(tbl.Kind.Neq(lifelong)).
	//	Or(tbl.Kind.Eq(lifelong), tbl.Tag.NotRegxp(ipv4)).
	//	Group(tbl.Tag).
	//	Order(tbl.Tag).
	//	Limit(1000).
	//	UnderlyingDB().
	//	Select("COUNT(*) AS count", "minion_tag.tag AS name").
	//	Scan(&ret)

	// SELECT COUNT(*) AS count, minion_tag.tag AS name
	// FROM minion_tag
	// WHERE minion_tag.kind <> 1
	//   OR (minion_tag.kind = 1 AND NOT minion_tag.tag ~ '^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$')
	// GROUP BY minion_tag.tag
	// ORDER BY minion_tag.tag
	// LIMIT 1000;
	str := `SELECT COUNT(*) AS count, minion_tag.tag AS name
FROM minion_tag
WHERE minion_tag.kind <> 1
   OR (minion_tag.kind = 1 AND NOT minion_tag.tag ~ $1)
GROUP BY minion_tag.tag
ORDER BY minion_tag.tag
LIMIT 1000
`
	_ = tbl.WithContext(ctx).
		UnderlyingDB().
		Raw(str, ipv4).
		Scan(&ret)

	return ret
}
