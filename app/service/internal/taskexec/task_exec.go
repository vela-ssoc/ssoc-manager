package taskexec

import (
	"context"
	"encoding/json"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/vela-ssoc/vela-manager/bridge/linkhub"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

func New(qry *query.Query, hub linkhub.Huber) *TaskExec {
	return &TaskExec{
		qry: qry,
		hub: hub,
	}
}

type TaskExec struct {
	qry *query.Query
	hub linkhub.Huber
}

func (te *TaskExec) Exec(ctx context.Context, taskID int64) error {
	// 查询 task 并将环境数据拷贝到 task_execute 库中用作快照
	// 根据 filter 和 exclude 条件匹配所有的符合条件的节点生成到 item 表中
	// 推送器开始依次推送任务

	return te.doExec(ctx, taskID)
}

func (te *TaskExec) Count(ctx context.Context, filters model.TaskExecuteFilter, excludes []string) (int64, error) {
	return te.qry.Minion.WithContext(ctx).Where(te.whereSQL(ctx, filters, excludes)...).Count()
}

func (te *TaskExec) whereSQL(ctx context.Context, f model.TaskExecuteFilter, excludes []string) []gen.Condition {
	minion, minionTag := te.qry.Minion, te.qry.MinionTag
	minionDo, minionTagDo := minion.WithContext(ctx), minionTag.WithContext(ctx)

	deleted := uint8(model.MSDelete)
	wheres := []gen.Condition{minion.Status.Neq(deleted)}
	if f.InetMode {
		if len(f.Inets) != 0 {
			wheres = append(wheres, minion.Inet.In(f.Inets...))
		}
	} else {
		var inputs request.CondWhereInputs
		for _, filter := range f.Filters {
			inputs.Filters = append(inputs.Filters, &request.CondWhereInput{
				Key:      filter.Key,
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		}

		args := &request.KeywordConditions{
			Keywords: request.Keywords{Keyword: f.Keyword},
			Conditions: request.Conditions{
				CondWhereInputs: inputs,
			},
		}
		_ = args
		// conds, _ := te.svc.CompileWhere(args)
		// wheres = append(wheres, conds...)
	}

	if len(excludes) != 0 {
		excludeSQL := minionTagDo.Select(minionTag.MinionID).Where(minionTag.Tag.In(excludes...))
		expr := minionTagDo.Columns(minionTag.MinionID).NotIn(excludeSQL)
		wheres = append(wheres, expr)
	}
	expr := minionDo.Columns(minion.ID).In(minionTagDo.Select(minionTag.MinionID).Where(wheres...))

	return append(wheres, expr)
}

func (*TaskExec) inetMode(inets, excludes []string) []gen.Condition {
	return nil
}

func (te *TaskExec) doExec(ctx context.Context, taskID int64) error {
	startedAt := time.Now()
	var finished bool
	var execID int64
	status := model.TaskExecuteStatus{CreatedAt: startedAt, UpdatedAt: startedAt}
	brokerIDs := make(map[int64]int, 8)
	err := te.qry.Transaction(func(tx *query.Query) error {
		taskExtension, minion := tx.TaskExtension, tx.Minion
		taskExecute, taskExecuteItem := tx.TaskExecute, tx.TaskExecuteItem
		taskExtensionDo := taskExtension.WithContext(ctx)
		taskExecuteDo, taskExecuteItemDo := taskExecute.WithContext(ctx), taskExecuteItem.WithContext(ctx)
		task, err := taskExtensionDo.Where(taskExtension.ID.Eq(taskID)).First()
		if err != nil {
			return err
		}

		wheres := te.whereSQL(ctx, task.Filters, task.Excludes)
		minionDo := minion.WithContext(ctx).Where(wheres...)

		total, err := minionDo.Count()
		if err != nil {
			return err
		}

		status.Total = int(total)
		finished = total == 0
		execute := &model.TaskExecute{
			TaskID:        taskID,
			Name:          task.Name,
			Intro:         task.Intro,
			Status:        status,
			Finished:      finished,
			Code:          task.Code,
			CodeSHA1:      task.CodeSHA1,
			ContentQuote:  task.ContentQuote,
			Cron:          task.Cron,
			SpecificTimes: task.SpecificTimes,
			Timeout:       task.Timeout,
			PushSize:      task.PushSize,
			Filters:       task.Filters,
			Excludes:      task.Excludes,
			CreatedBy:     task.CreatedBy,
			UpdatedBy:     task.UpdatedBy,
			CreatedAt:     startedAt,
			UpdatedAt:     startedAt,
		}
		if err = taskExecuteDo.Create(execute); err != nil {
			return err
		}

		timeout := task.Timeout.Duration()
		if timeout <= time.Minute {
			timeout = 3 * time.Minute
		} else if timeout > time.Hour {
			timeout = 10*time.Minute + timeout
		} else {
			timeout = 2 * timeout
		}

		execID = execute.ID
		const batchSize = 500
		var buf []*model.Minion
		if err = minionDo.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
			now := time.Now()
			expiredAt := now.Add(timeout)

			items := make([]*model.TaskExecuteItem, 0, batchSize)
			for _, m := range buf {
				brokerIDs[m.BrokerID] += 1
				item := &model.TaskExecuteItem{
					TaskID:     taskID,
					ExecID:     execID,
					MinionID:   m.ID,
					Inet:       m.Inet,
					BrokerID:   m.BrokerID,
					BrokerName: m.BrokerName,
					ExpiredAt:  expiredAt,
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				items = append(items, item)
			}

			return taskExecuteItemDo.Create(items...)
		}); err != nil {
			return err
		}

		_, err = taskExtensionDo.Where(taskExtension.ID.Eq(taskID)).
			UpdateSimple(taskExtension.Status.Value(status), taskExtension.Finished.Value(false), taskExtension.ExecID.Value(execID))

		return err
	})
	if err != nil || finished {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	const rawURL = "/api/v1/task/push"
	type requestData struct {
		ExecID int64 `json:"exec_id"`
	}

	var failedN int
	taskExecuteItem := te.qry.TaskExecuteItem
	for bid, cnt := range brokerIDs {
		err = te.hub.Oneway(ctx, bid, rawURL, &requestData{ExecID: execID})

		taskExecuteItemDo := taskExecuteItem.WithContext(ctx)
		wheres := []gen.Condition{taskExecuteItem.TaskID.Eq(taskID), taskExecuteItem.BrokerID.Eq(bid)}
		var updates []field.AssignExpr
		mStatus := &model.TaskStepStatus{Succeed: err == nil, ExecutedAt: time.Now()}
		if err != nil {
			failedN += cnt
			mStatus.Reason = err.Error()
			updates = append(updates,
				taskExecuteItem.Finished.Value(true),
				taskExecuteItem.ErrorCode.Value(model.TaskExecuteErrorCodeBroker),
			)
		}
		updates = append(updates, taskExecuteItem.ManagerStatus.Value(mStatus))
		_, _ = taskExecuteItemDo.Where(wheres...).UpdateSimple(updates...)
	}

	finished = failedN >= status.Total
	status.Failed, status.UpdatedAt = failedN, time.Now()
	taskExtension := te.qry.TaskExtension
	taskExtension.WithContext(ctx).Where(taskExtension.ID.Eq(taskID)).UpdateSimple(taskExtension.Status.Value(status), taskExtension.Finished.Value(finished))

	taskExecute := te.qry.TaskExecute
	taskExecute.WithContext(ctx).Where(taskExecute.ID.Eq(execID)).UpdateSimple(taskExecute.Status.Value(status), taskExecute.Finished.Value(finished))

	if !finished {
		go te.watchResult(24*time.Hour, taskID, execID, status)
	}

	return nil
}

func (te *TaskExec) watchResult(timeout time.Duration, taskID, execID int64, status model.TaskExecuteStatus) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	startedAt := time.Now()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			status.UpdatedAt = now
			if te.scanResult(ctx, taskID, execID, status) {
				return
			}

			if sub := now.Sub(startedAt); sub > time.Minute {
				ticker.Reset(15 * time.Second)
			} else if sub > 10*time.Minute {
				ticker.Reset(30 * time.Second)
			}
		}
	}
}

func (te *TaskExec) scanResult(ctx context.Context, taskID, execID int64, status model.TaskExecuteStatus) bool {
	type countData struct {
		Succeed int `gorm:"column:succeed"`
		Failed  int `gorm:"column:failed"`
		Total   int `gorm:"column:total"`
	}
	rawSQL := "SELECT COUNT(CASE WHEN finished = TRUE AND succeed = TRUE THEN TRUE END)  AS succeed," +
		"    COUNT(CASE WHEN finished = TRUE AND succeed = FALSE THEN TRUE END) AS failed," +
		"    COUNT(*)                                                           AS total " +
		"FROM task_execute_item " +
		"WHERE exec_id = ?"
	data := new(countData)
	taskExecuteItem := te.qry.TaskExecuteItem
	err := taskExecuteItem.WithContext(ctx).
		UnderlyingDB().
		Raw(rawSQL, execID).
		Scan(data).
		Error
	if err != nil {
		return false
	}

	status.Total = data.Total
	status.Succeed = data.Succeed
	status.Failed = data.Failed
	finished := data.Succeed+data.Failed >= data.Total

	taskExecute := te.qry.TaskExecute
	taskExtension := te.qry.TaskExtension
	taskExtension.WithContext(ctx).
		Where(taskExtension.ID.Eq(taskID), taskExtension.ExecID.Eq(execID)).
		UpdateSimple(taskExtension.Status.Value(status), taskExtension.Finished.Value(finished))
	taskExecute.WithContext(ctx).Where(taskExecute.ID.Eq(execID), taskExecute.TaskID.Eq(taskID)).
		UpdateSimple(taskExecute.Status.Value(status), taskExecute.Finished.Value(finished))

	return finished
}

// TaskExecuteData 任务执行时 broker 向 agent 下发的任务内容。
type TaskExecuteData struct {
	// ID 即任务的唯一标识。
	ID int64 `json:"id"`

	// ExecID 任务的执行 ID。
	// 同一个任务可以被多次触发执行，每次执行时都会生成一个新的 ExecID，
	// 用于标识任务执行的不同批次。
	ExecID int64 `json:"exec_id"`

	// Name 任务名，比如：内网 Log4J 扫描。
	// 注意：当前中心端要求任务 Name 唯一，但是 agent 尽量不要直接拿 Name 区分唯一性，
	// 一是业务可能会变化，二是任务可以删除新建，名字可能相同。
	// 唯一性判断请以 ID 为准。
	Name string `json:"name"`

	// Intro 任务简介给人看的，对程序处理来说无实际意义。
	Intro string `json:"intro"`

	// Code 可运行的 Lua 代码。
	Code string `json:"code"`

	// CodeSHA1 Lua 代码的 SHA-1 值（小写）。
	CodeSHA1 string `json:"code_sha1"`

	// Timeout 任务超时时间。
	// 注意：该时间可能为 0 （即：未指定超时时间），对此 agent 可自行
	Timeout time.Duration `json:"timeout"`
}

// TaskExecuteResult 任务执行完毕后 agent 向 broker 发送的回执。
// agent 执行任务完后主动向 broker 发起一个 http 请求。
type TaskExecuteResult struct {
	ID      int64           `json:"id"`      // 任务 ID
	ExecID  int64           `json:"exec_id"` // 执行 ID
	Succeed int             `json:"succeed"` // 是否执行成功
	Result  json.RawMessage `json:"result"`  // 成功结果（如果有的话）
}
