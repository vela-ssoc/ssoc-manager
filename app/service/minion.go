package service

import (
	"context"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
)

type MinionService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.MinionSummary)
	Detail(ctx context.Context, id int64) (*param.MinionDetail, error)
}

func Minion() MinionService {
	return &minionService{}
}

type minionService struct{}

func (mon *minionService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*param.MinionSummary) {
	tagTbl := query.MinionTag
	monTbl := query.Minion

	db := monTbl.WithContext(ctx).UnderlyingDB().
		Table("minion").
		Distinct("minion.id").
		Joins("LEFT JOIN minion_tag ON minion.id = minion_tag.minion_id").
		Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}
	var monIDs []int64
	if db.Order("minion.id").
		Scopes(page.DBScope(count)).
		Scan(&monIDs); len(monIDs) == 0 {
		return 0, nil
	}
	// 查询数据
	minions, err := monTbl.WithContext(ctx).Where(monTbl.ID.In(monIDs...)).Find()
	if err != nil {
		return 0, nil
	}

	tagMap := map[int64][]string{}
	infoMap := map[int64]*model.SysInfo{}

	if tags, _ := tagTbl.WithContext(ctx).
		Where(tagTbl.MinionID.In(monIDs...)).
		Find(); len(tags) != 0 {
		tagMap = model.MinionTags(tags).ToMap()
	}
	infoTbl := query.SysInfo
	if infos, _ := infoTbl.WithContext(ctx).Where(infoTbl.ID.In(monIDs...)).Find(); len(infos) != 0 {
		infoMap = model.SysInfos(infos).ToMap()
	}

	ret := make([]*param.MinionSummary, 0, len(monIDs))
	for _, m := range minions {
		id := m.ID
		ms := &param.MinionSummary{
			ID:      id,
			Inet:    m.Inet,
			Goos:    m.Goos,
			Edition: m.Edition,
			Status:  m.Status,
			IDC:     m.IDC,
			IBu:     m.IBu,
			Comment: m.Comment,
			Tags:    tagMap[id],
		}
		if ms.Tags == nil {
			ms.Tags = []string{}
		}
		if inf := infoMap[id]; inf != nil {
			ms.CPUCore = inf.CPUCore
			ms.MemTotal = inf.MemTotal
			ms.MemFree = inf.MemFree
		}
		ret = append(ret, ms)
	}

	return count, ret
}

func (mon *minionService) Detail(ctx context.Context, id int64) (*param.MinionDetail, error) {
	monTbl := query.Minion
	infoTbl := query.SysInfo
	dat := new(param.MinionDetail)
	err := monTbl.WithContext(ctx).
		LeftJoin(infoTbl, infoTbl.ID.EqCol(monTbl.ID)).
		Where(monTbl.ID.Eq(id)).
		Scan(&dat)
	if err != nil {
		return nil, err
	}

	tagTbl := query.MinionTag
	dat.Tags, _ = tagTbl.WithContext(ctx).Where(tagTbl.MinionID.Eq(id)).Find()
	if dat.Tags == nil {
		dat.Tags = []*model.MinionTag{}
	}

	return dat, nil
}
