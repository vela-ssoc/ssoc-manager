package service

import (
	"context"
	"strconv"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type MinionTaskService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionTask)
	Detail(ctx context.Context, mid, sid int64) (*param.MinionTaskDetail, error)
	Minion(ctx context.Context, mid int64) ([]*param.MinionTaskSummary, error)
}

func MinionTask() MinionTaskService {
	return &minionTaskService{}
}

type minionTaskService struct{}

func (biz *minionTaskService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, []*model.MinionTask) {
	db := query.MinionTask.WithContext(ctx).UnderlyingDB().Scopes(scope.Where)
	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.MinionTask
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}

func (biz *minionTaskService) Detail(ctx context.Context, mid, sid int64) (*param.MinionTaskDetail, error) {
	subTbl := query.Substance
	sub, err := subTbl.WithContext(ctx).Where(subTbl.ID.Eq(sid)).First()
	if err != nil {
		return nil, err
	}
	if sub.MinionID != 0 && sub.MinionID != mid {
		return nil, errcode.ErrSubstanceNotExist
	}

	taskTbl := query.MinionTask
	mt, _ := taskTbl.WithContext(ctx).
		Where(taskTbl.MinionID.Eq(mid)).
		Where(taskTbl.SubstanceID.Eq(sid)).
		First()

	dialect := sub.MinionID == mid
	dat := &param.MinionTaskDetail{
		ID:         sid,
		Name:       sub.Name,
		Icon:       sub.Icon,
		From:       mt.From,
		Status:     mt.Status,
		Link:       mt.Link,
		Desc:       sub.Desc,
		Dialect:    dialect,
		LegalHash:  sub.Hash,
		ActualHash: mt.Hash,
		Failed:     mt.Failed,
		Cause:      mt.Cause,
		Chunk:      sub.Chunk,
		Version:    sub.Version,
		CreatedAt:  sub.CreatedAt,
		UpdatedAt:  sub.UpdatedAt,
		TaskAt:     mt.CreatedAt,
		Uptime:     mt.Uptime,
		Runners:    mt.Runners,
	}

	return dat, nil
}

func (biz *minionTaskService) Minion(ctx context.Context, mid int64) ([]*param.MinionTaskSummary, error) {
	monTbl := query.Minion
	mon, err := monTbl.WithContext(ctx).Select(monTbl.Inet).Where(monTbl.ID.Eq(mid)).First()
	if err != nil {
		return nil, errcode.ErrNodeNotExist
	}

	tagTbl := query.MinionTag
	effTbl := query.Effect

	// SELECT * FROM effect WHERE enable = true AND tag IN (SELECT DISTINCT tag FROM minion_tag WHERE minion_id = $mid)
	subSQL := tagTbl.WithContext(ctx).Distinct(tagTbl.Tag).Where(tagTbl.MinionID.Eq(mid))
	effs, err := effTbl.WithContext(ctx).
		Where(effTbl.Enable.Is(true)).
		Where(effTbl.WithContext(ctx).Columns(effTbl.Tag).In(subSQL)).
		Find()
	comIDs, subIDs := model.Effects(effs).Exclusion(mon.Inet)
	if len(comIDs) != 0 {
		comTbl := query.Compound
		coms, err := comTbl.WithContext(ctx).Select(comTbl.ID).Where(comTbl.ID.In(comIDs...)).Find()
		if err == nil {
			ids := model.Compounds(coms).SubstanceIDs()
			subIDs = append(subIDs, ids...)
		}
	}

	subTbl := query.Substance
	subs, err := subTbl.WithContext(ctx).
		Omit(subTbl.Chunk).
		Where(subTbl.MinionID.Eq(mid)).
		Or(subTbl.ID.In(subIDs...)).
		Find()
	if err != nil {
		return nil, err
	}

	taskTbl := query.MinionTask
	mts, _ := taskTbl.WithContext(ctx).Where(taskTbl.MinionID.Eq(mid)).Find()

	taskMap := make(map[string]*model.MinionTask, len(mts))
	for _, mt := range mts {
		// 注意：从 console 加载的的配置脚本是没有 SubstanceID 的，要做特殊处理
		var subID string
		if mt.SubstanceID > 0 {
			subID = strconv.FormatInt(mt.SubstanceID, 10)
		} else {
			subID = strconv.FormatInt(mt.ID, 10) + mt.Name
		}
		taskMap[subID] = mt
	}

	// 上报的数据与数据库数据合并整理
	res := make([]*param.MinionTaskSummary, 0, len(subs)+8)
	for _, sub := range subs {
		dialect := sub.MinionID == mid

		tv := &param.MinionTaskSummary{
			ID: sub.ID, Name: sub.Name, Icon: sub.Icon, Dialect: dialect,
			LegalHash: sub.Hash, CreatedAt: sub.CreatedAt, UpdatedAt: sub.UpdatedAt,
		}

		subID := strconv.FormatInt(sub.ID, 10)
		task := taskMap[subID]
		if task != nil {
			delete(taskMap, subID)
			tv.From, tv.Status, tv.Link, tv.ActualHash = task.From, task.Status, task.Link, task.Hash
		}
		res = append(res, tv)
	}

	for _, task := range taskMap {
		tv := &param.MinionTaskSummary{
			Name: task.Name, From: task.From, Status: task.Status, Link: task.Link,
			ActualHash: task.Hash, CreatedAt: task.CreatedAt, UpdatedAt: task.CreatedAt,
		}
		res = append(res, tv)
	}

	return res, nil
}
