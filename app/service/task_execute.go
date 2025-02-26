package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/param/response"
	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/vela-ssoc/vela-manager/param/mrequest"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

type TaskExecute struct {
	qry *query.Query
	log *slog.Logger
}

func NewTaskExecute(qry *query.Query, log *slog.Logger) *TaskExecute {
	return &TaskExecute{
		qry: qry,
		log: log,
	}
}

func (tex *TaskExecute) Page(ctx context.Context, args *mrequest.TaskExecutePage) (*response.Pages[*model.TaskExecute], error) {
	tbl := tex.qry.TaskExecute
	var wheres []gen.Condition
	if taskID := args.TaskID; taskID > 0 {
		wheres = append(wheres, tbl.TaskID.Eq(taskID))
	}
	dao := tbl.WithContext(ctx).Where(wheres...).
		Scopes(args.LikeScopes(true, tbl.Name, tbl.Intro))

	page := response.NewPages[*model.TaskExecute](args.Page, args.Size)
	cnt, err := dao.Count()
	if err != nil {
		return nil, err
	} else if cnt == 0 {
		return page.Empty(), nil
	}

	records, err := dao.Order(tbl.ID.Desc()).
		Scopes(page.FP(cnt)).
		Find()
	if err != nil {
		return nil, err
	}

	return page.SetRecords(records), nil
}

func (tex *TaskExecute) Remove(ctx context.Context, id int64) error {
	tbl := tex.qry.TaskExecute
	dat, err := tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).First()
	if err != nil {
		return err
	}
	if !dat.Finished {
		return errcode.ErrTaskBusy
	}

	return tex.qry.Transaction(func(tx *query.Query) error {
		item := tx.TaskExecuteItem
		if _, err = item.WithContext(ctx).
			Where(item.TaskID.Eq(id)).
			Delete(); err != nil {
			return err
		}

		exec := tx.TaskExecute
		_, err = exec.WithContext(ctx).
			Where(exec.ID.Eq(id)).
			Delete()

		return err
	})
}

// TimeoutMonitor 监控超时任务。
func (tex *TaskExecute) TimeoutMonitor(ctx context.Context) error {
	now := time.Now()
	itemTbl := tex.qry.TaskExecuteItem

	status := &model.TaskStepStatus{
		Succeed:    false,
		Reason:     "任务执行超时",
		ExecutedAt: now,
	}

	execIDs := make(map[int64]struct{}, 16)
	var items []*model.TaskExecuteItem
	_ = itemTbl.WithContext(ctx).
		Where(itemTbl.ExpiredAt.Lt(now), itemTbl.Finished.Is(false)).
		FindInBatches(&items, 100, func(tx gen.Dao, _ int) error {
			for _, item := range items {
				execIDs[item.ExecID] = struct{}{}

				updates := []field.AssignExpr{
					itemTbl.Succeed.Value(false),
					itemTbl.Finished.Value(true),
					itemTbl.MinionStatus.Value(status),
				}
				_, _ = tx.Where(itemTbl.ID.Eq(item.ID)).
					UpdateSimple(updates...)
			}

			return nil
		})

	return tex.executeMonitor(ctx, execIDs)
}

func (tex *TaskExecute) executeMonitor(ctx context.Context, execIDs map[int64]struct{}) error {
	for execID := range execIDs {
		status, err := tex.reStatus(ctx, execID)
		if err != nil {
			continue
		}

		execTbl := tex.qry.TaskExecute
		execTbl.WithContext(ctx).
			Where(execTbl.ID.Eq(execID)).
			UpdateSimple(execTbl.Finished.Value(status.Finished()), execTbl.Status.Value(status))

		extTbl := tex.qry.TaskExtension
		extTbl.WithContext(ctx).
			Where(extTbl.ExecID.Eq(execID)).
			UpdateSimple(extTbl.Finished.Value(status.Finished()), extTbl.Status.Value(status))
	}

	return nil
}

func (tex *TaskExecute) reStatus(ctx context.Context, execID int64) (*model.TaskStatus, error) {
	rawSQL := "SELECT COUNT(CASE WHEN finished = TRUE AND succeed = TRUE THEN TRUE END)  AS succeed," +
		"    COUNT(CASE WHEN finished = TRUE AND succeed = FALSE THEN TRUE END) AS failed," +
		"    COUNT(*)                                                           AS total " +
		"FROM task_execute_item " +
		"WHERE exec_id = ?"
	data := new(model.TaskStatus)
	taskExecuteItem := tex.qry.TaskExecuteItem
	err := taskExecuteItem.WithContext(ctx).
		UnderlyingDB().
		Raw(rawSQL, execID).
		Scan(data).
		Error

	return data, err
}
