package service

import (
	"context"
	"log/slog"
	"strconv"
	"sync"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/ssoc-manager/param/mresponse"
	"gorm.io/gorm"
)

func NewMinionTask(qry *query.Query, log *slog.Logger) *MinionTask {
	return &MinionTask{
		qry: qry,
		log: log,
	}
}

type MinionTask struct {
	qry *query.Query
	log *slog.Logger
}

func (mt *MinionTask) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, param.TaskList) {
	db := mt.qry.MinionTask.WithContext(ctx).UnderlyingDB()
	stmt := db.Table("minion_task AS minion_task").
		Joins("LEFT JOIN substance st ON st.id = minion_task.substance_id").
		Scopes(scope.Where).
		Order("minion_task.id DESC")

	var count int64
	if stmt.Count(&count); count == 0 {
		return 0, nil
	}

	var dats param.TaskList
	stmt.Select("minion_task.id", "minion_task.inet", "minion_task.minion_id", "minion_task.substance_id",
		"minion_task.name", "minion_task.dialect", "minion_task.status", "minion_task.hash AS report_hash", "st.hash",
		"minion_task.link", "minion_task.from", "minion_task.failed", "minion_task.cause", "st.created_at",
		"st.updated_at", "minion_task.created_at AS report_at").
		Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}

func (mt *MinionTask) Detail(ctx context.Context, mid, sid int64) (*param.MinionTaskDetail, error) {
	subTbl := mt.qry.Substance
	sub, err := subTbl.WithContext(ctx).Where(subTbl.ID.Eq(sid)).First()
	if err != nil {
		return nil, err
	}
	if sub.MinionID != 0 && sub.MinionID != mid {
		return nil, errcode.ErrSubstanceNotExist
	}

	taskTbl := mt.qry.MinionTask
	task, _ := taskTbl.WithContext(ctx).
		Where(taskTbl.MinionID.Eq(mid), taskTbl.SubstanceID.Eq(sid)).
		Order(taskTbl.ID.Desc()).
		First()
	if task == nil {
		task = new(model.MinionTask)
	}

	dialect := sub.MinionID == mid
	dat := &param.MinionTaskDetail{
		ID:           sid,
		Name:         sub.Name,
		Icon:         sub.Icon,
		From:         task.From,
		Status:       task.Status,
		Link:         task.Link,
		Desc:         sub.Desc,
		Dialect:      dialect,
		LegalHash:    sub.Hash,
		ActualHash:   task.Hash,
		Failed:       task.Failed,
		Cause:        task.Cause,
		Chunk:        sub.Chunk,
		ContentQuote: sub.ContentQuote,
		Version:      sub.Version,
		CreatedAt:    sub.CreatedAt,
		UpdatedAt:    sub.UpdatedAt,
		TaskAt:       task.CreatedAt,
		Uptime:       task.Uptime,
		Runners:      task.Runners,
	}

	return dat, nil
}

func (mt *MinionTask) Minion(ctx context.Context, mid int64) ([]*param.MinionTaskSummary, error) {
	monTbl := mt.qry.Minion
	tagTbl := mt.qry.MinionTag
	effTbl := mt.qry.Effect
	taskTbl := mt.qry.MinionTask

	mon, err := monTbl.WithContext(ctx).
		Select(monTbl.Inet, monTbl.Unload).
		Where(monTbl.ID.Eq(mid)).
		First()
	if err != nil {
		return nil, errcode.ErrNodeNotExist
	}

	var subs []*model.Substance
	if !mon.Unload {
		// SELECT * FROM effect WHERE enable = true AND tag IN (SELECT DISTINCT tag FROM minion_tag WHERE minion_id = $mid)
		subSQL := tagTbl.WithContext(ctx).Distinct(tagTbl.Tag).Where(tagTbl.MinionID.Eq(mid))
		effs, _ := effTbl.WithContext(ctx).
			Where(effTbl.Enable.Is(true)).
			Where(effTbl.WithContext(ctx).Columns(effTbl.Tag).In(subSQL)).
			Find()
		subIDs := model.Effects(effs).Exclusion(mon.Inet)
		subTbl := mt.qry.Substance
		dats, exx := subTbl.WithContext(ctx).
			Omit(subTbl.Chunk).
			Where(subTbl.MinionID.Eq(mid)).
			Or(subTbl.ID.In(subIDs...)).
			Find()
		if exx != nil {
			return nil, exx
		}
		subs = dats
	}

	mts, _ := taskTbl.WithContext(ctx).Where(taskTbl.MinionID.Eq(mid)).Find()

	taskMap := make(map[string]*model.MinionTask, len(mts))
	for _, m := range mts {
		// 注意：从 console 加载的的配置脚本是没有 SubstanceID 的，要做特殊处理
		var subID string
		if m.SubstanceID > 0 {
			subID = strconv.FormatInt(m.SubstanceID, 10)
		} else {
			subID = strconv.FormatInt(m.ID, 10) + m.Name
		}
		taskMap[subID] = m
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
			Name: task.Name, From: task.From, Status: task.Status, Link: task.Link, Dialect: task.Dialect,
			ActualHash: task.Hash, CreatedAt: task.CreatedAt, UpdatedAt: task.CreatedAt,
		}
		res = append(res, tv)
	}

	return res, nil
}

func (mt *MinionTask) Gather(ctx context.Context, page param.Pager) (int64, []*param.TaskGather) {
	db := mt.qry.MinionTask.WithContext(ctx).UnderlyingDB()
	ctSQL := db.Model(&model.MinionTask{}).
		Select("name", "COUNT(*) count").
		Group("name")
	if kw := page.Keyword(); kw != "" {
		ctSQL.Where("name LIKE ?", kw)
	}
	var count int64
	if ctSQL.Count(&count); count == 0 {
		return 0, nil
	}

	var cts request.NameCounts
	ctSQL.Order("count DESC").Scopes(page.DBScope(count)).Scan(&cts)
	if len(cts) == 0 {
		return 0, nil
	}

	mutex := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	ret := make([]*param.TaskGather, len(cts))
	for i, ct := range cts {
		wg.Add(1)
		go mt.gather(wg, mutex, db, ct.Name, i, ret)
	}
	wg.Wait()

	return count, ret
}

func (mt *MinionTask) gather(wg *sync.WaitGroup, mutex *sync.Mutex, db *gorm.DB, name string, n int, ret []*param.TaskGather) {
	defer wg.Done()

	rawSQL := "SELECT COUNT(IF(dialect = TRUE, TRUE, NULL)) AS dialect, " +
		"COUNT(IF(dialect = FALSE, TRUE, NULL))    AS public, " +
		"COUNT(IF(status = 'running', TRUE, NULL)) AS running," +
		"COUNT(IF(status = 'doing', TRUE, NULL))   AS doing, " +
		"COUNT(IF(status = 'fail', TRUE, NULL))    AS fail, " +
		"COUNT(IF(status = 'panic', TRUE, NULL))   AS panic, " +
		"COUNT(IF(status = 'reg', TRUE, NULL))     AS reg, " +
		"COUNT(IF(status = 'update', TRUE, NULL))  AS update " +
		"FROM minion_task WHERE name = ?"

	var res param.TaskCount
	db.Raw(rawSQL, name).Scan(&res)
	tg := &param.TaskGather{
		Name:    name,
		Dialect: res.Dialect > 0,
		Running: res.Running,
		Doing:   res.Doing,
		Fail:    res.Fail,
		Panic:   res.Panic,
		Reg:     res.Reg,
		Update:  res.Update,
	}

	mutex.Lock()
	defer mutex.Unlock()
	ret[n] = tg
}

func (mt *MinionTask) Count(ctx context.Context) *param.TaskCount {
	rawSQL := "SELECT COUNT(IF(dialect = TRUE, TRUE, NULL)) AS dialect, " +
		"COUNT(IF(dialect = FALSE, TRUE, NULL))    AS public, " +
		"COUNT(IF(status = 'running', TRUE, NULL)) AS running," +
		"COUNT(IF(status = 'doing', TRUE, NULL))   AS doing, " +
		"COUNT(IF(status = 'fail', TRUE, NULL))    AS fail, " +
		"COUNT(IF(status = 'panic', TRUE, NULL))   AS panic, " +
		"COUNT(IF(status = 'reg', TRUE, NULL))     AS reg, " +
		"COUNT(IF(status = 'update', TRUE, NULL))  AS update " +
		"FROM minion_task"

	res := new(param.TaskCount)
	db := mt.qry.MinionTask.WithContext(ctx).UnderlyingDB()
	db.Raw(rawSQL).Scan(&res)

	return res
}

func (mt *MinionTask) RCount(ctx context.Context, pager param.Pager) (int64, []*param.TaskRCount) {
	size := pager.Size()
	ret := make([]*param.TaskRCount, 0, size)

	tbl := mt.qry.MinionTask
	count, _ := tbl.WithContext(ctx).
		Distinct(tbl.SubstanceID).
		Where(tbl.SubstanceID.Neq(0)).
		Count()
	if count == 0 {
		return 0, ret
	}

	strSQL := "SELECT substance_id AS id, COUNT(*) AS count " +
		" FROM minion_task " +
		" WHERE substance_id != 0 " +
		" GROUP BY substance_id " +
		" ORDER BY count " +
		" DESC LIMIT ?, ? "
	mt.qry.MinionTask.
		WithContext(ctx).
		UnderlyingDB().
		Scopes(pager.DBScope(count)).
		Raw(strSQL, 0, size).
		Scan(&ret)

	index := make(map[int64]*param.TaskRCount, size)
	sids := make([]int64, 0, size)
	for _, rc := range ret {
		sid := rc.ID
		if _, ok := index[sid]; ok {
			continue
		}
		index[sid] = rc
		sids = append(sids, sid)
	}
	if len(sids) != 0 {
		stbl := mt.qry.Substance
		subs, _ := stbl.WithContext(ctx).
			Select(stbl.ID, stbl.Name, stbl.Desc).
			Where(stbl.ID.In(sids...)).
			Find()
		for _, sub := range subs {
			sid := sub.ID
			if rc := index[sid]; rc != nil {
				rc.Name = sub.Name
				rc.Desc = sub.Desc
			}
		}
	}

	return count, ret
}

func (mt *MinionTask) Tasks(ctx context.Context, minionID int64) (*mresponse.MinionTask, error) {
	// 查询 minionID 相关的配置
	// 查询 minionID 排除的

	monTbl := mt.qry.Minion
	monDao := monTbl.WithContext(ctx)
	taskTbl := mt.qry.MinionTask
	taskDao := taskTbl.WithContext(ctx)
	excTbl := mt.qry.MinionSubstanceExclude
	excDao := excTbl.WithContext(ctx)

	minion, err := monDao.Where(monTbl.ID.Eq(minionID)).First()
	if err != nil {
		return nil, err
	}
	// 1.0 通过 IP 简单排除
	subs, excludeInets, err := mt.substances(ctx, minionID, minion.Inet)
	if err != nil {
		return nil, err
	}

	ret := &mresponse.MinionTask{Unload: minion.Unload}
	reportTasks, _ := taskDao.Where(taskTbl.MinionID.Eq(minionID)).Find()
	excludeTasks, _ := excDao.Where(excTbl.MinionID.Eq(minionID)).Find() // 2.0 排除

	excludes := make(map[int64]bool, 4)
	for _, tsk := range excludeTasks {
		excludes[tsk.SubstanceID] = true
	}
	reportMaps := make(map[int64]*mresponse.MinionTaskItemReport, 8)
	for _, tsk := range reportTasks {
		reportMaps[tsk.SubstanceID] = &mresponse.MinionTaskItemReport{
			From:      tsk.From,
			Uptime:    tsk.Uptime,
			Link:      tsk.Link,
			Status:    tsk.Status,
			Hash:      tsk.Hash,
			Cause:     tsk.Cause,
			Runners:   tsk.Runners,
			CreatedAt: tsk.CreatedAt,
		}
	}
	for _, sub := range subs {
		id := sub.ID
		report := reportMaps[id]
		delete(reportMaps, id)
		tsk := &mresponse.MinionTaskItem{
			ID:           id,
			Name:         sub.Name,
			Icon:         sub.Icon,
			Dialect:      sub.MinionID == minionID,
			Excluded:     excludes[id],
			ExcludedInet: excludeInets[id],
			Desc:         sub.Desc,
			Hash:         sub.Hash,
			Report:       report,
			ContentQuote: sub.ContentQuote,
			CreatedAt:    sub.CreatedAt,
			UpdatedAt:    sub.UpdatedAt,
		}
		ret.Tasks = append(ret.Tasks, tsk)
	}
	for _, task := range reportTasks {
		subID := task.SubstanceID
		report := reportMaps[subID]
		if report == nil {
			continue
		}
		tsk := &mresponse.MinionTaskItem{
			Name: task.Name,
			Report: &mresponse.MinionTaskItemReport{
				From:      task.From,
				Uptime:    task.Uptime,
				Link:      task.Link,
				Status:    task.Status,
				CreatedAt: task.CreatedAt,
			},
		}
		ret.Tasks = append(ret.Tasks, tsk)
	}

	return ret, nil
}

func (mt *MinionTask) Task(ctx context.Context, minionID, substanceID int64) (*mresponse.MinionTaskItem, error) {
	subTbl := mt.qry.Substance
	subDao := subTbl.WithContext(ctx)
	sub, err := subDao.Where(subTbl.ID.Eq(substanceID)).First()
	if err != nil {
		return nil, err
	}
	if sub.MinionID != 0 && sub.MinionID != minionID {
		return nil, nil
	}

	item := &mresponse.MinionTaskItem{
		ID:           sub.ID,
		Name:         sub.Name,
		Icon:         sub.Icon,
		Dialect:      sub.MinionID == minionID,
		Hash:         sub.Hash,
		Desc:         sub.Desc,
		Chunk:        sub.Chunk,
		Version:      sub.Version,
		ContentQuote: sub.ContentQuote,
		CreatedAt:    sub.CreatedAt,
		UpdatedAt:    sub.UpdatedAt,
	}

	taskTbl := mt.qry.MinionTask
	task, _ := taskTbl.WithContext(ctx).
		Where(taskTbl.MinionID.Eq(minionID), taskTbl.SubstanceID.Eq(substanceID)).First()
	if task != nil {
		item.Report = &mresponse.MinionTaskItemReport{
			From:      task.From,
			Uptime:    task.Uptime,
			Link:      task.Link,
			Status:    task.Status,
			Hash:      task.Hash,
			Cause:     task.Cause,
			Runners:   task.Runners,
			CreatedAt: task.CreatedAt,
		}
	}
	// 检查该策略是否被排除（新版）
	excTbl := mt.qry.MinionSubstanceExclude
	if cnt, _ := excTbl.WithContext(ctx).
		Where(excTbl.MinionID.Eq(minionID), excTbl.SubstanceID.Eq(substanceID)).
		Count(); cnt != 0 {
		item.Excluded = true
	}
	// TODO 检查该策略是否被排除（旧版）

	return item, nil
}

func (mt *MinionTask) substances(ctx context.Context, minionID int64, inet string) (model.Substances, map[int64]bool, error) {
	// 查询该节点所有相关的标签
	tagTbl := mt.qry.MinionTag
	tagDao := tagTbl.WithContext(ctx)
	effTbl := mt.qry.Effect
	effDao := effTbl.WithContext(ctx)
	subTbl := mt.qry.Substance
	subDao := subTbl.WithContext(ctx)

	minionTags, _ := tagDao.Where(tagTbl.MinionID.Eq(minionID)).Find()

	tags := make([]string, 0, len(minionTags))
	for _, tag := range minionTags {
		tags = append(tags, tag.Tag)
	}

	excludes := make(map[int64]bool, 8)
	subIDs := make([]int64, 0, 10)
	if len(tags) != 0 {
		effects, _ := effDao.Where(effTbl.Enable.Is(true), effTbl.Tag.In(tags...)).Find()
		for _, eff := range effects {
			subID := eff.EffectID
			subIDs = append(subIDs, subID)
			for _, v := range eff.Exclusion {
				if v == inet {
					excludes[subID] = true
					break
				}
			}
		}
	}

	dao := subDao.Where(subTbl.MinionID.Eq(minionID))
	if len(subIDs) != 0 {
		dao.Or(subTbl.ID.In(subIDs...))
	}
	subs, err := dao.Find()
	if err != nil {
		return nil, nil, err
	}

	return subs, excludes, nil
}
