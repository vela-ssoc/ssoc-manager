package service

import (
	"context"
	"sync"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type EffectService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*param.EffectSummary)
	Create(ctx context.Context, ec *param.EffectCreate, userID int64) (int64, error)
	Update(ctx context.Context, eu *param.EffectUpdate, userID int64) (int64, error)
	Delete(ctx context.Context, submitID int64) (int64, error)
	Progress(ctx context.Context, tid int64) *param.EffectProgress
	Progresses(ctx context.Context, tid int64, page param.Pager) (int64, []*model.SubstanceTask)
}

func Effect(pusher push.Pusher, seq SequenceService, task SubstanceTaskService) EffectService {
	return &effectService{
		pusher: pusher,
		seq:    seq,
		task:   task,
	}
}

type effectService struct {
	pusher push.Pusher
	seq    SequenceService
	task   SubstanceTaskService
	mutex  sync.RWMutex
}

func (eff *effectService) Page(ctx context.Context, page param.Pager) (int64, []*param.EffectSummary) {
	effTbl := query.Effect
	dao := effTbl.WithContext(ctx).Distinct(effTbl.SubmitID)
	if kw := page.Keyword(); kw != "" {
		dao.Where(effTbl.Name.Like(kw))
	}

	count, err := dao.Count()
	if err != nil || count == 0 {
		return 0, nil
	}
	size := page.Size()
	submitIDs := make([]int64, 0, size)
	_ = dao.Order(effTbl.SubmitID).Scopes(page.Scope(count)).Scan(&submitIDs)
	if len(submitIDs) == 0 {
		return 0, nil
	}

	effs, err := effTbl.WithContext(ctx).
		Where(effTbl.SubmitID.In(submitIDs...)).
		Order(effTbl.SubmitID).
		Find()
	if err != nil || len(effs) == 0 {
		return 0, nil
	}

	tagMap := make(map[int64]map[string]struct{}, size)
	idx := make(map[int64]*param.EffectSummary, size)
	ret := make([]*param.EffectSummary, 0, size)

	subIDs := make([]int64, 0, 20)
	subMap := make(map[int64]struct{}, 20)
	effSubMap := make(map[int64]map[int64]struct{}, 16)

	for _, e := range effs {
		id, tag, eid := e.SubmitID, e.Tag, e.EffectID
		sm, ok := idx[id]
		if !ok {
			sm = &param.EffectSummary{
				ID:         id,
				Name:       e.Name,
				Tags:       make([]string, 0, 10),
				Enable:     e.Enable,
				Version:    e.Version,
				Exclusion:  e.Exclusion,
				Compounds:  make([]*param.IDName, 0, 10),
				Substances: make([]*param.IDName, 0, 10),
				CreatedAt:  e.CreatedAt,
				UpdatedAt:  e.UpdatedAt,
			}
			idx[id] = sm
			ret = append(ret, sm)
			tagMap[id] = make(map[string]struct{}, 8)
		}
		if _, exist := tagMap[id][tag]; !exist {
			tagMap[id][tag] = struct{}{}
			sm.Tags = append(sm.Tags, tag)
		}

		if effSubMap[id] == nil {
			effSubMap[id] = make(map[int64]struct{}, 8)
		}
		if _, exist := effSubMap[id][eid]; !exist {
			effSubMap[id][eid] = struct{}{}
			sm.Substances = append(sm.Substances, &param.IDName{ID: eid})
		}

		if _, exist := subMap[eid]; !exist {
			subMap[eid] = struct{}{}
			subIDs = append(subIDs, eid)
		}
	}

	comKV := make(map[int64]string, 16)
	subKV := make(map[int64]string, 16)

	if len(subIDs) != 0 {
		subTbl := query.Substance
		subs, _ := subTbl.WithContext(ctx).
			Select(subTbl.ID, subTbl.Name).
			Where(subTbl.ID.In(subIDs...)).
			Find()
		for _, s := range subs {
			subKV[s.ID] = s.Name
		}
	}

	for _, sm := range ret {
		for _, c := range sm.Compounds {
			c.Name = comKV[c.ID]
		}
		for _, s := range sm.Substances {
			s.Name = subKV[s.ID]
		}
	}

	return count, ret
}

func (eff *effectService) Create(ctx context.Context, ec *param.EffectCreate, userID int64) (int64, error) {
	eff.mutex.Lock()
	defer eff.mutex.Unlock()

	// 名字不能重复
	name := ec.Name
	tbl := query.Effect
	count, err := tbl.WithContext(ctx).Where(tbl.Name.Eq(name)).Count()
	if err != nil || count != 0 {
		return 0, errcode.FmtErrNameExist.Fmt(name)
	}

	if err = ec.Check(ctx); err != nil {
		return 0, err
	}
	if ec.Enable {
		if err = eff.task.BusyError(ctx); err != nil {
			return 0, err
		}
	}

	submitID := eff.seq.Generate()
	effects := ec.Expand(submitID, userID)
	// 插入数据库
	if err = tbl.WithContext(ctx).
		CreateInBatches(effects, 200); err != nil || !ec.Enable {
		return 0, err
	}

	return eff.task.AsyncTags(ctx, ec.Tags)
}

func (eff *effectService) Update(ctx context.Context, eu *param.EffectUpdate, userID int64) (int64, error) {
	eff.mutex.Lock()
	defer eff.mutex.Unlock()

	// 查询原有数据
	subID, version := eu.ID, eu.Version
	tbl := query.Effect
	effs, err := tbl.WithContext(ctx).
		Where(tbl.SubmitID.Eq(subID)).
		Where(tbl.Version.Eq(version)).
		Find()
	if err != nil || len(effs) == 0 {
		return 0, errcode.ErrVersion
	}
	reduce := model.Effects(effs).Reduce()

	if name := eu.Name; reduce.Name != name { // 修改名字则检查名字不可重复
		count, err := tbl.WithContext(ctx).Where(tbl.Name.Eq(name)).Count()
		if err != nil || count != 0 {
			return 0, errcode.FmtErrNameExist.Fmt(name)
		}
	}

	if err = eu.Check(ctx); err != nil {
		return 0, err
	}

	// 根据前端提交的数据，分为以下三种情况：
	// 一、无需下发任务更新通知：
	// 		1. Exclusion、Tags、Substances、Enable 这四个字段均未修改。
	//		2. 修改前是关闭状态，提交的还是关闭状态，无论修改其它什么字段，都不会下发更新任务。
	// 二、轻量级更新：
	// 		1. 本次提交只修改了 Exclusion 字段，并且提交前和本次提交的 Enable 都是 true。
	// 三、全量更新：
	// 		1. 修改 Tags、Substances、Enable 这几个字段的任意一个值就需要全量更新。
	var heavy, light bool
	nothing := !eu.Enable && !reduce.Enable
	if !nothing {
		heavy = !eff.equalsStrings(eu.Tags, reduce.Tags) ||
			!eff.equalsInt64s(eu.Substances, reduce.Substances)
		if !heavy {
			light = !heavy && !eff.equalsStrings(eu.Exclusion, reduce.Exclusion)
		}
	}

	if !nothing { // 无论轻量级还是重量级的修改都要加锁
		if err = eff.task.BusyError(ctx); err != nil {
			return 0, err
		}
	}

	effects := eu.Expand(reduce, userID)
	err = query.Q.Transaction(func(tx *query.Query) error {
		dao := tx.Effect.WithContext(ctx)
		res, err := dao.Where(tbl.SubmitID.Eq(subID)).Where(tbl.Version.Eq(version)).Delete()
		if err != nil {
			return err
		}
		if res.RowsAffected == 0 {
			return errcode.ErrVersion
		}
		return dao.CreateInBatches(effects, 200)
	})
	if err != nil || nothing {
		return 0, err
	}

	if light {
		diffs := eff.diff(eu.Exclusion, reduce.Exclusion)
		return eff.task.AsyncInets(ctx, diffs)
	}

	allTags := eff.mergeStrings(eu.Tags, reduce.Tags)
	return eff.task.AsyncTags(ctx, allTags)
}

func (eff *effectService) Delete(ctx context.Context, submitID int64) (int64, error) {
	eff.mutex.Lock()
	defer eff.mutex.Unlock()

	tbl := query.Effect
	effs, err := tbl.WithContext(ctx).
		Where(tbl.SubmitID.Eq(submitID)).
		Find()
	if err != nil || len(effs) == 0 {
		return 0, err
	}

	if err = eff.task.BusyError(ctx); err != nil {
		return 0, err
	}

	reduce := model.Effects(effs).Reduce()
	if _, err = tbl.WithContext(ctx).
		Where(tbl.SubmitID.Eq(submitID)).
		Delete(); err != nil || !reduce.Enable {
		// 未启用的配置就是未下发的，删除后可以不通知节点
		return 0, err
	}

	return eff.task.AsyncTags(ctx, reduce.Tags)
}

// Progress 任务进度
func (eff *effectService) Progress(ctx context.Context, tid int64) *param.EffectProgress {
	return eff.task.Progress(ctx, tid)
}

// Progresses 获取当前最后一次运行的任务信息
func (eff *effectService) Progresses(ctx context.Context, tid int64, page param.Pager) (int64, []*model.SubstanceTask) {
	return eff.task.Progresses(ctx, tid, page)
}

func (*effectService) equalsStrings(as, bs []string) bool {
	size := len(as)
	if size != len(bs) {
		return false
	}
	hm := make(map[string]struct{}, size)
	for _, s := range as {
		hm[s] = struct{}{}
	}
	for _, s := range bs {
		if _, ok := hm[s]; !ok {
			return false
		}
		delete(hm, s)
	}

	return len(hm) == 0
}

func (*effectService) equalsInt64s(as, bs []int64) bool {
	size := len(as)
	if size != len(bs) {
		return false
	}
	hm := make(map[int64]struct{}, size)
	for _, s := range as {
		hm[s] = struct{}{}
	}
	for _, s := range bs {
		if _, ok := hm[s]; !ok {
			return false
		}
		delete(hm, s)
	}

	return len(hm) == 0
}

func (*effectService) mergeStrings(as, bs []string) []string {
	size := len(as) + len(bs)
	ret := make([]string, 0, size)
	hm := make(map[string]struct{}, size)

	for _, s := range as {
		if _, ok := hm[s]; !ok {
			hm[s] = struct{}{}
			ret = append(ret, s)
		}
	}
	for _, s := range bs {
		if _, ok := hm[s]; !ok {
			hm[s] = struct{}{}
			ret = append(ret, s)
		}
	}

	return ret
}

func (*effectService) diff(as, bs []string) []string {
	hm := make(map[string]struct{}, len(as))
	for _, s := range as {
		hm[s] = struct{}{}
	}

	ret := make([]string, 0, 16)
	for _, s := range bs {
		if _, ok := hm[s]; ok {
			delete(hm, s)
		} else {
			ret = append(ret, s)
		}
	}
	for s := range hm {
		ret = append(ret, s)
	}

	return ret
}
