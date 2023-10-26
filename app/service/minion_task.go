package service

import (
	"context"
	"strconv"
	"sync"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/errcode"
	"gorm.io/gorm"
)

type MinionTaskService interface {
	Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, param.TaskList)
	Detail(ctx context.Context, mid, sid int64) (*param.MinionTaskDetail, error)
	Minion(ctx context.Context, mid int64) ([]*param.MinionTaskSummary, error)
	Gather(ctx context.Context, page param.Pager) (int64, []*param.TaskGather)
	Count(ctx context.Context) *param.TaskCount
	RCount(ctx context.Context, pager param.Pager) (int64, []*param.TaskRCount)
}

func MinionTask() MinionTaskService {
	return &minionTaskService{}
}

type minionTaskService struct{}

func (biz *minionTaskService) Page(ctx context.Context, page param.Pager, scope dynsql.Scope) (int64, param.TaskList) {
	db := query.MinionTask.WithContext(ctx).UnderlyingDB()
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
		"minion_task.link", "minion_task.`from`", "minion_task.failed", "minion_task.cause", "st.created_at",
		"st.updated_at", "minion_task.created_at AS report_at").
		Scopes(page.DBScope(count)).Find(&dats)

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
		Where(taskTbl.MinionID.Eq(mid), taskTbl.SubstanceID.Eq(sid)).
		Order(taskTbl.ID.Desc()).
		First()
	if mt == nil {
		mt = new(model.MinionTask)
	}

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
	tagTbl := query.MinionTag
	effTbl := query.Effect
	taskTbl := query.MinionTask

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
		if len(subIDs) != 0 {
			subTbl := query.Substance
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
	}

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

func (biz *minionTaskService) Gather(ctx context.Context, page param.Pager) (int64, []*param.TaskGather) {
	db := query.MinionTask.WithContext(ctx).UnderlyingDB()
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

	var cts []*param.NameCount
	ctSQL.Order("count DESC").Scopes(page.DBScope(count)).Scan(&cts)
	if len(cts) == 0 {
		return 0, nil
	}

	mutex := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	ret := make([]*param.TaskGather, len(cts))
	for i, ct := range cts {
		wg.Add(1)
		go biz.gather(wg, mutex, db, ct.Name, i, ret)
	}
	wg.Wait()

	return count, ret
}

func (biz *minionTaskService) gather(wg *sync.WaitGroup, mutex *sync.Mutex, db *gorm.DB, name string, n int, ret []*param.TaskGather) {
	defer wg.Done()

	rawSQL := "SELECT COUNT(IF(`dialect` = TRUE, TRUE, NULL)) AS `dialect`, " +
		"COUNT(IF(`dialect` = FALSE, TRUE, NULL))    AS `public`, " +
		"COUNT(IF(`status` = 'running', TRUE, NULL)) AS `running`," +
		"COUNT(IF(`status` = 'doing', TRUE, NULL))   AS `doing`, " +
		"COUNT(IF(`status` = 'fail', TRUE, NULL))    AS `fail`, " +
		"COUNT(IF(`status` = 'panic', TRUE, NULL))   AS `panic`, " +
		"COUNT(IF(`status` = 'reg', TRUE, NULL))     AS `reg`, " +
		"COUNT(IF(`status` = 'update', TRUE, NULL))  AS `update` " +
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

func (biz *minionTaskService) Count(ctx context.Context) *param.TaskCount {
	rawSQL := "SELECT COUNT(IF(`dialect` = TRUE, TRUE, NULL)) AS `dialect`, " +
		"COUNT(IF(`dialect` = FALSE, TRUE, NULL))    AS `public`, " +
		"COUNT(IF(`status` = 'running', TRUE, NULL)) AS `running`," +
		"COUNT(IF(`status` = 'doing', TRUE, NULL))   AS `doing`, " +
		"COUNT(IF(`status` = 'fail', TRUE, NULL))    AS `fail`, " +
		"COUNT(IF(`status` = 'panic', TRUE, NULL))   AS `panic`, " +
		"COUNT(IF(`status` = 'reg', TRUE, NULL))     AS `reg`, " +
		"COUNT(IF(`status` = 'update', TRUE, NULL))  AS `update` " +
		"FROM minion_task"

	res := new(param.TaskCount)
	db := query.MinionTask.WithContext(ctx).UnderlyingDB()
	db.Raw(rawSQL).Scan(&res)

	return res
}

func (biz *minionTaskService) RCount(ctx context.Context, pager param.Pager) (int64, []*param.TaskRCount) {
	size := pager.Size()
	ret := make([]*param.TaskRCount, 0, size)

	tbl := query.MinionTask
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
	query.MinionTask.
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
		stbl := query.Substance
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
