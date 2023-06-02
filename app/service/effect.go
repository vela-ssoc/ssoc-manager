package service

import (
	"context"
	"sync"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/app/internal/transact"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
)

type EffectService interface {
	Page(ctx context.Context, page param.Pager) (int64, []*param.EffectSummary)
	Create(ctx context.Context, ec *param.EffectCreate, userID int64) (int64, error)
	Update(ctx context.Context, eu *param.EffectUpdate, userID int64) (int64, error)
	Delete(ctx context.Context, submitID int64) (int64, error)
}

func Effect(pusher push.Pusher, seq SequenceService) EffectService {
	return &effectService{
		pusher:  pusher,
		timeout: time.Hour,
		seq:     seq,
	}
}

type effectService struct {
	pusher  push.Pusher
	seq     SequenceService
	timeout time.Duration
	taskID  int64
	mutex   sync.RWMutex
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
	comIDs := make([]int64, 0, 20)
	comMap := make(map[int64]struct{}, 20)
	effSubMap := make(map[int64]map[int64]struct{}, 16)
	effComMap := make(map[int64]map[int64]struct{}, 16)

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
		if e.Compound {
			if effComMap[id] == nil {
				effComMap[id] = make(map[int64]struct{}, 8)
			}
			if _, exist := effComMap[id][eid]; !exist {
				effComMap[id][eid] = struct{}{}
				sm.Compounds = append(sm.Compounds, &param.IDName{ID: eid})
			}

			if _, exist := comMap[eid]; !exist {
				comMap[eid] = struct{}{}
				comIDs = append(comIDs, eid)
			}
		} else {
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
	}

	comKV := make(map[int64]string, 16)
	subKV := make(map[int64]string, 16)
	if len(comIDs) != 0 {
		comTbl := query.Compound
		compounds, _ := comTbl.WithContext(ctx).
			Select(comTbl.ID, comTbl.Name).
			Where(comTbl.ID.In(comIDs...)).
			Find()
		for _, c := range compounds {
			comKV[c.ID] = c.Name
		}
	}
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
	if ec.Enable && eff.existTask(ctx) {
		return 0, errcode.ErrTaskBusy
	}

	submitID := eff.seq.Generate()
	effects := ec.Expand(submitID, userID)
	// 插入数据库
	if err = tbl.WithContext(ctx).
		CreateInBatches(effects, 200); err != nil {
		return 0, err
	}

	var taskID int64
	if ec.Enable {
		taskID = eff.seq.Generate()
		brokerIDs, err := transact.EffectTaskTx(ctx, taskID, ec.Tags)
		if err != nil {
			return 0, err
		}
		eff.taskID = taskID

		// 推送任务
		eff.pusher.TaskTable(ctx, brokerIDs, taskID)
	}

	return taskID, nil
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
	if err != nil {
		return 0, err
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

	task := eu.Enable != reduce.Enable ||
		!eff.equalsStrings(eu.Tags, reduce.Tags) ||
		!eff.equalsInt64s(eu.Substances, reduce.Substances) ||
		!eff.equalsInt64s(eu.Compounds, reduce.Compounds) ||
		!eff.equalsStrings(eu.Tags, reduce.Tags)
	if task && eff.existTask(ctx) {
		return 0, errcode.ErrTaskBusy
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
	if err != nil {
		return 0, err
	}

	var taskID int64
	if task {
		allTags := eff.mergeStrings(eu.Tags, reduce.Tags)
		taskID = eff.seq.Generate()
		brokerIDs, err := transact.EffectTaskTx(ctx, taskID, allTags)
		if err != nil {
			return 0, err
		}
		eff.taskID = taskID

		// 推送任务
		eff.pusher.TaskTable(ctx, brokerIDs, taskID)
	}

	return taskID, nil
}

func (eff *effectService) Delete(ctx context.Context, submitID int64) (int64, error) {
	eff.mutex.Lock()
	defer eff.mutex.Unlock()

	tbl := query.Effect
	effs, err := tbl.WithContext(ctx).Where(tbl.SubmitID.Eq(submitID)).Find()
	if err != nil {
		return 0, err
	}

	if eff.existTask(ctx) {
		return 0, errcode.ErrTaskBusy
	}

	reduce := model.Effects(effs).Reduce()
	if _, err = tbl.WithContext(ctx).Where(tbl.SubmitID.Eq(submitID)).Delete(); err != nil {
		return 0, err
	}

	var taskID int64
	if reduce.Enable {
		taskID = eff.seq.Generate()
		brkIDs, err := transact.EffectTaskTx(ctx, taskID, reduce.Tags)
		if err != nil {
			return 0, err
		}
		eff.taskID = taskID

		// 推送任务
		eff.pusher.TaskTable(ctx, brkIDs, taskID)
	}

	return taskID, nil
}

// Progress 获取当前最后一次运行的任务信息
func (eff *effectService) Progress(ctx context.Context) {
	// 查询当前任务
}

// hasRunning 是否有正在运行的任务
func (eff *effectService) existTask(ctx context.Context) bool {
	if eff.taskID == 0 {
		return false
	}

	before := time.Now().Add(-eff.timeout)
	tbl := query.SubstanceTask
	count, err := tbl.WithContext(ctx).
		Where(tbl.TaskID.Eq(eff.taskID)).
		Where(tbl.Executed.Is(false)).
		Where(tbl.CreatedAt.Gte(before)).
		Count()

	return err == nil && count != 0
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
