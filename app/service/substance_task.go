package service

import (
	"context"
	"sync"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/dynsql"
	"github.com/vela-ssoc/vela-manager/app/internal/param"
	"github.com/vela-ssoc/vela-manager/bridge/push"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/vela-ssoc/vela-manager/param/mrequest"
	"github.com/vela-ssoc/vela-manager/param/mresponse"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

type SubstanceTaskService interface {
	AsyncTags(ctx context.Context, tags []string) (int64, error)
	AsyncInets(ctx context.Context, inets []string) (int64, error)
	Progress(ctx context.Context, tid int64) *mresponse.EffectProgress
	Progresses(ctx context.Context, tid int64, page mrequest.Pager) (int64, []*model.SubstanceTask)
	BusyError(ctx context.Context) error

	Page(ctx context.Context, id int64, page param.Pager, scope dynsql.Scope, likes []gen.Condition) (int64, []*model.SubstanceTask)
	Histories(ctx context.Context, page param.Pager, scope dynsql.Scope, likes []gen.Condition) (int64, []*model.SubstanceTask)
}

func SubstanceTask(qry *query.Query, seq SequenceService, pusher push.Pusher) SubstanceTaskService {
	return &substanceTaskService{
		qry:     qry,
		seq:     seq,
		pusher:  pusher,
		timeout: time.Hour,
	}
}

type substanceTaskService struct {
	qry     *query.Query
	seq     SequenceService
	pusher  push.Pusher
	timeout time.Duration
	mutex   sync.Mutex
}

func (biz *substanceTaskService) Page(ctx context.Context, id int64, page param.Pager, scope dynsql.Scope, likes []gen.Condition) (int64, []*model.SubstanceTask) {
	if id == 0 {
		id = biz.currentTaskID(ctx)
	}
	if id == 0 {
		return 0, nil
	}

	return biz.page(ctx, id, page, scope, likes)
}

func (biz *substanceTaskService) Histories(ctx context.Context, page param.Pager, scope dynsql.Scope, likes []gen.Condition) (int64, []*model.SubstanceTask) {
	return biz.page(ctx, 0, page, scope, likes)
}

func (biz *substanceTaskService) page(ctx context.Context, id int64, page param.Pager, scope dynsql.Scope, likes []gen.Condition) (int64, []*model.SubstanceTask) {
	tbl := biz.qry.SubstanceTask
	dao := tbl.WithContext(ctx).
		Order(tbl.TaskID.Desc(), tbl.ID)
	if id != 0 {
		dao.Where(tbl.TaskID.Eq(id))
	}

	if len(likes) != 0 {
		for i, like := range likes {
			likes[i] = dao.Or(like)
		}
		dao.Where(likes...)
	}
	db := dao.UnderlyingDB().Scopes(scope.Where)

	var count int64
	if db.Count(&count); count == 0 {
		return 0, nil
	}

	var dats []*model.SubstanceTask
	db.Scopes(page.DBScope(count)).Find(&dats)

	return count, dats
}

func (biz *substanceTaskService) AsyncTags(ctx context.Context, tags []string) (int64, error) {
	if len(tags) == 0 {
		return 0, nil
	}

	biz.mutex.Lock()
	defer biz.mutex.Unlock()
	if err := biz.BusyError(ctx); err != nil {
		return 0, err
	}

	tid := biz.seq.Generate()
	go biz.insertTagTask(tid, tags)

	return tid, nil
}

func (biz *substanceTaskService) AsyncInets(ctx context.Context, inets []string) (int64, error) {
	if len(inets) == 0 {
		return 0, nil
	}

	biz.mutex.Lock()
	defer biz.mutex.Unlock()
	if err := biz.BusyError(ctx); err != nil {
		return 0, err
	}

	tid := biz.seq.Generate()
	go biz.insertInetTasks(tid, inets)

	return tid, nil
}

func (biz *substanceTaskService) Progress(ctx context.Context, tid int64) *mresponse.EffectProgress {
	if tid <= 0 {
		tid = biz.currentTaskID(ctx)
	}
	if tid <= 0 {
		return new(mresponse.EffectProgress)
	}

	return biz.progress(ctx, tid)
}

func (biz *substanceTaskService) progress(ctx context.Context, id int64) *mresponse.EffectProgress {
	taskDo := biz.qry.SubstanceTask.WithContext(ctx)
	db := taskDo.UnderlyingDB()
	dialect := db.Dialector.Name()
	if dialect == "postgres" || dialect == "opengauss" {
		return biz.progressForOpenGauss(db, id)
	} else {
		return biz.progressForMySQL(db, id)
	}
}

func (biz *substanceTaskService) progressForOpenGauss(db *gorm.DB, id int64) *mresponse.EffectProgress {
	ret := new(mresponse.EffectProgress)
	strSQL := "SELECT COUNT(*)                      AS count, " +
		"COUNT(IF(executed, TRUE, NULL))            AS executed, " +
		"COUNT(IF(executed AND failed, TRUE, NULL)) AS failed " +
		"FROM substance_task " +
		"WHERE task_id = ?"
	db.Raw(strSQL, id).Scan(ret)

	return ret
}

func (biz *substanceTaskService) progressForMySQL(db *gorm.DB, id int64) *mresponse.EffectProgress {
	ret := new(mresponse.EffectProgress)
	strSQL := `
SELECT COUNT(*)                                                    AS count,
	      SUM(CASE WHEN executed THEN TRUE ELSE FALSE END)            AS executed,
	      SUM(CASE WHEN executed AND failed THEN TRUE ELSE FALSE END) AS failed
	FROM substance_task
	WHERE task_id = ?`
	db.Raw(strSQL, id).Scan(ret)

	return ret
}

// Progresses 获取当前最后一次运行的任务信息
func (biz *substanceTaskService) Progresses(ctx context.Context, tid int64, page mrequest.Pager) (int64, []*model.SubstanceTask) {
	if tid <= 0 {
		tid = biz.currentTaskID(ctx)
	}
	if tid <= 0 {
		return 0, nil
	}

	tbl := biz.qry.SubstanceTask
	dao := tbl.WithContext(ctx)
	cond := []gen.Condition{
		dao.Where(tbl.TaskID.Eq(tid)),
	}
	if kw := page.Keyword(); kw != "" {
		like := dao.Or(tbl.Inet.Like(kw)).
			Or(tbl.BrokerName.Like(kw)).
			Or(tbl.Reason.Like(kw))
		cond = append(cond, like)
	}

	dao = dao.Where(cond...)
	count, _ := dao.Count()
	if count == 0 {
		return 0, nil
	}
	dats, _ := dao.Scopes(page.Scope(count)).Find()

	return count, dats
}

func (biz *substanceTaskService) BusyError(ctx context.Context) error {
	// 超时控制
	now := time.Now()
	timeout := 10 * time.Minute
	tbl := biz.qry.SubstanceTask
	wheres := []gen.Condition{
		tbl.Executed.Is(false),
		tbl.CreatedAt.Lt(now.Add(-timeout)),
	}
	updates := []field.AssignExpr{
		tbl.Executed.Value(true),
		tbl.Failed.Value(true),
		tbl.Reason.Value("任务下发超时"),
	}
	_, _ = tbl.WithContext(ctx).
		Where(wheres...).
		UpdateSimple(updates...)

	if tid := biz.currentTaskID(ctx); tid != 0 {
		return errcode.FmtErrTaskBusy.Fmt(tid)
	}

	return nil
}

func (biz *substanceTaskService) currentTaskID(ctx context.Context) int64 {
	cat := time.Now().Add(-biz.timeout)
	tbl := biz.qry.SubstanceTask
	tsk, err := tbl.WithContext(ctx).
		Where(tbl.CreatedAt.Gte(cat), tbl.Executed.Is(false)).
		Order(tbl.ID.Desc()).
		First()
	if err == nil {
		return tsk.TaskID
	}
	return 0
}

func (biz *substanceTaskService) insertTagTask(tid int64, tags []string) {
	ctx, cancel := context.WithTimeout(context.Background(), biz.timeout)
	defer cancel()

	// 删除久远的任务，防止任务表的数据越来越多
	du := 5 * biz.timeout
	if du < 7*24*time.Hour {
		du = 7 * 24 * time.Hour
	}
	tbl := biz.qry.SubstanceTask
	_, _ = tbl.WithContext(ctx).
		Where(tbl.CreatedAt.Lt(time.Now().Add(-du))).
		Delete()

	// 查询相关节点
	limit := 500
	var offsetID int64
	tagTbl := biz.qry.MinionTag
	monTbl := biz.qry.Minion
	bmap := make(map[int64]struct{}, 16)

	for {
		var minionIDs []int64
		err := tagTbl.WithContext(ctx).
			Distinct(tagTbl.MinionID).
			Where(tagTbl.MinionID.Gt(offsetID), tagTbl.Tag.In(tags...)).
			Order(tagTbl.MinionID).
			Limit(limit).
			Scan(&minionIDs)
		qsz := len(minionIDs)
		if err != nil || qsz == 0 {
			break
		}
		offsetID = minionIDs[qsz-1]

		// 查询节点信息
		nodes, _ := monTbl.WithContext(ctx).
			Select(monTbl.ID, monTbl.Status, monTbl.Inet, monTbl.BrokerID, monTbl.BrokerName).
			Where(monTbl.ID.In(minionIDs...)).
			Find()
		nsz := len(nodes)
		if nsz == 0 {
			continue
		}

		tasks := make([]*model.SubstanceTask, 0, nsz)
		for _, node := range nodes {
			bid := node.BrokerID
			if bid == 0 || node.Status == model.MSDelete {
				continue
			}
			task := &model.SubstanceTask{
				TaskID:     tid,
				MinionID:   node.ID,
				Inet:       node.Inet,
				BrokerID:   bid,
				BrokerName: node.BrokerName,
			}
			tasks = append(tasks, task)
			bmap[bid] = struct{}{}
		}

		if len(tasks) == 0 {
			continue
		}
		_ = tbl.WithContext(ctx).Create(tasks...)
	}

	bids := make([]int64, 0, 16)
	for id := range bmap {
		bids = append(bids, id)
	}

	if len(bids) != 0 {
		biz.pusher.TaskTable(ctx, bids, tid)
	}
}

func (biz *substanceTaskService) insertInetTasks(tid int64, inets []string) {
	ctx, cancel := context.WithTimeout(context.Background(), biz.timeout)
	defer cancel()

	tbl := biz.qry.Minion
	minions, err := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.Inet, tbl.BrokerID, tbl.BrokerName).
		Where(tbl.Inet.In(inets...)).
		Find()
	if err != nil || len(minions) == 0 {
		return
	}

	bids, _ := biz.insertInetTask(ctx, tid, minions)
	biz.pusher.TaskTable(ctx, bids, tid)
}

func (biz *substanceTaskService) insertInetTask(ctx context.Context, tid int64, minions []*model.Minion) ([]int64, error) {
	hm := make(map[int64]struct{}, 16)
	size := len(minions)
	if size == 0 {
		return nil, nil
	}

	rows := make([]*model.SubstanceTask, 0, len(minions))
	for _, m := range minions {
		mid, bid := m.ID, m.BrokerID
		if m.BrokerID == 0 || m.Status == model.MSDelete {
			continue
		}
		task := &model.SubstanceTask{
			TaskID:     tid,
			MinionID:   mid,
			Inet:       m.Inet,
			BrokerID:   bid,
			BrokerName: m.BrokerName,
		}
		rows = append(rows, task)
		hm[bid] = struct{}{}
	}

	tbl := biz.qry.SubstanceTask
	err := tbl.WithContext(ctx).Create(rows...)

	ret := make([]int64, 0, len(hm))
	for id := range hm {
		ret = append(ret, id)
	}

	return ret, err
}
