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
	Progress(ctx context.Context, tid int64) *param.EffectProgress
	Progresses(ctx context.Context, tid int64, page param.Pager) (int64, []*model.SubstanceTask)
}

func Effect(pusher push.Pusher, seq SequenceService) EffectService {
	return &effectService{
		pusher:  pusher,
		timeout: 10 * time.Minute,
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
	if ec.Enable && eff.existTask(ctx) {
		return 0, errcode.ErrTaskBusy
	}

	submitID := eff.seq.Generate()
	effects := ec.Expand(submitID, userID)
	// 插入数据库
	if err = tbl.WithContext(ctx).
		CreateInBatches(effects, 200); err != nil || !ec.Enable {
		return 0, err
	}

	taskID := eff.seq.Generate()
	eff.taskID = taskID

	go func() {
		brkIDs, exx := transact.EffectTaskTx(ctx, taskID, ec.Tags)
		if exx == nil {
			eff.pusher.TaskTable(ctx, brkIDs, taskID)
		}
	}()

	return taskID, nil
}

func (eff *effectService) Update(ctx context.Context, eu *param.EffectUpdate, userID int64) (int64, error) {
	eff.mutex.Lock()
	defer eff.mutex.Unlock()

	// 查询任务
	if eff.existTask(ctx) {
		return 0, errcode.ErrTaskBusy
	}

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

	task := eu.Enable != reduce.Enable ||
		!eff.equalsStrings(eu.Tags, reduce.Tags) ||
		!eff.equalsInt64s(eu.Substances, reduce.Substances) ||
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
	if err != nil || !task {
		return 0, err
	}

	taskID := eff.seq.Generate()
	eff.taskID = taskID
	allTags := eff.mergeStrings(eu.Tags, reduce.Tags)

	go func() {
		brkIDs, exx := transact.EffectTaskTx(ctx, taskID, allTags)
		if exx == nil {
			eff.pusher.TaskTable(ctx, brkIDs, taskID)
		}
	}()

	return taskID, nil
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

	if eff.existTask(ctx) {
		return 0, errcode.ErrTaskBusy
	}

	reduce := model.Effects(effs).Reduce()
	if _, err = tbl.WithContext(ctx).
		Where(tbl.SubmitID.Eq(submitID)).
		Delete(); err != nil || !reduce.Enable {
		// 未启用的配置就是未下发的，删除后可以不通知节点
		return 0, err
	}

	taskID := eff.seq.Generate()
	eff.taskID = taskID

	go func() {
		brkIDs, exx := transact.EffectTaskTx(ctx, taskID, reduce.Tags)
		if exx == nil {
			eff.pusher.TaskTable(ctx, brkIDs, taskID)
		}
	}()

	return taskID, nil
}

// Progress 任务进度
func (eff *effectService) Progress(ctx context.Context, tid int64) *param.EffectProgress {
	if tid == 0 { // task == 0 就查询当前任务
		eff.mutex.Lock()
		tid = eff.taskID
		eff.mutex.Unlock()
	}

	ret := &param.EffectProgress{ID: tid}
	if tid == 0 {
		return ret
	}

	rawSQL := "SELECT COUNT(*)                      AS count, " +
		"COUNT(IF(executed, TRUE, NULL))            AS executed, " +
		"COUNT(IF(executed AND failed, TRUE, NULL)) AS failed " +
		"FROM substance_task " +
		"WHERE task_id = ?"
	db := query.SubstanceTask.
		WithContext(ctx).
		UnderlyingDB()
	db.Raw(rawSQL).Scan(ret)

	return ret
}

// Progresses 获取当前最后一次运行的任务信息
func (eff *effectService) Progresses(ctx context.Context, tid int64, page param.Pager) (int64, []*model.SubstanceTask) {
	if tid == 0 { // task == 0 就查询当前任务
		eff.mutex.Lock()
		tid = eff.taskID
		eff.mutex.Unlock()
	}
	if tid == 0 {
		return 0, nil
	}

	tbl := query.SubstanceTask
	dao := tbl.WithContext(ctx).
		Where(tbl.TaskID.Eq(tid))
	if kw := page.Keyword(); kw != "" {
		dao.Where(dao.Or(tbl.Inet.Like(kw), tbl.BrokerName.Like(kw), tbl.Reason.Like(kw)))
	}
	count, _ := dao.Count()
	if count == 0 {
		return 0, nil
	}
	dats, _ := dao.Scopes(page.Scope(count)).Find()

	return count, dats
}

// hasRunning 是否有正在运行的任务
func (eff *effectService) existTask(ctx context.Context) bool {
	tid := eff.taskID
	if tid == 0 {
		return false
	}

	before := time.Now().Add(-eff.timeout)
	tbl := query.SubstanceTask
	count, _ := tbl.WithContext(ctx).
		Where(tbl.TaskID.Eq(tid)).
		Where(tbl.Executed.Is(false)).
		Where(tbl.CreatedAt.Gte(before)).
		Count()

	return count != 0
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
