package service

import (
	"context"
	"strconv"

	"github.com/vela-ssoc/vela-common-mb/dal/dyncond"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/vela-ssoc/vela-common-mb/param/response"
	"github.com/vela-ssoc/vela-manager/param/mresponse"
	"gorm.io/gen/field"
)

func NewTaskExecuteItem(qry *query.Query) (*TaskExecuteItem, error) {
	tex := &TaskExecuteItem{qry: qry}

	opt := dyncond.Options{WhereCallback: tex.whereCallback()}
	tbl, err := dyncond.ParseModels(qry, []any{model.TaskExecuteItem{}}, opt)
	if err != nil {
		return nil, err
	}
	tex.tbl = tbl

	return tex, nil
}

type TaskExecuteItem struct {
	qry *query.Query
	tbl *dyncond.Tables
}

func (tex *TaskExecuteItem) Cond() *response.Cond {
	return response.ParseCond(tex.tbl)
}

func (tex *TaskExecuteItem) Page(ctx context.Context, args *request.PageConditions) (*response.Pages[*model.TaskExecuteItem], error) {
	wheres, _, err := tex.tbl.CompileWhere(args.CondWhereInputs.Inputs(), false)
	if err != nil {
		return nil, err
	}

	executeItem := tex.qry.TaskExecuteItem
	executeItemDo := executeItem.WithContext(ctx).Where(wheres...)

	pages := response.NewPages[*model.TaskExecuteItem](args.PageSize())
	cnt, err := executeItemDo.Count()
	if err != nil {
		return nil, err
	} else if cnt == 0 {
		return pages.Empty(), nil
	}

	items, err := executeItemDo.Scopes(pages.FP(cnt)).Find()
	if err != nil {
		return nil, err
	}

	return pages.SetRecords(items), nil
}

func (tex *TaskExecuteItem) CodeCounts(ctx context.Context, execID int64) (mresponse.TaskExecuteItemCodeCounts, error) {
	tbl := tex.qry.TaskExecuteItem
	ret := make(mresponse.TaskExecuteItemCodeCounts, 4)
	name, count := ret.Aliases()
	if err := tbl.WithContext(ctx).
		Select(tbl.ErrorCode.As(name), tbl.ErrorCode.Count().As(count)).
		Where(tbl.ExecID.Eq(execID)).
		Group(tbl.ErrorCode).
		Scan(&ret); err != nil {
		return nil, err
	}

	return ret, nil
}

func (tex *TaskExecuteItem) whereCallback() func(tbl *dyncond.Tables, w *dyncond.Where) *dyncond.Where {
	item := tex.qry.TaskExecuteItem
	ignores := []field.Expr{
		item.ManagerStatus, item.MinionStatus, item.BrokerStatus, item.Result,
	}

	return func(tbl *dyncond.Tables, w *dyncond.Where) *dyncond.Where {
		for _, ignore := range ignores {
			if tbl.EqualsExpr(ignore, w.Expr) {
				return nil
			}
		}
		if tbl.EqualsExpr(item.ErrorCode, w.Expr) {
			w.Enums = dyncond.Enums{
				{Key: strconv.FormatInt(model.TaskExecuteErrorCodeRunning, 10), Desc: "执行中"},
				{Key: strconv.FormatInt(model.TaskExecuteErrorCodeBroker, 10), Desc: "代理节点出错"},
				{Key: strconv.FormatInt(model.TaskExecuteErrorCodeAgent, 10), Desc: "Agent下发出错"},
				{Key: strconv.FormatInt(model.TaskExecuteErrorCodeExec, 10), Desc: "Agent执行出错"},
				{Key: strconv.FormatInt(model.TaskExecuteErrorCodeTimeout, 10), Desc: "Agent执行超时"},
				{Key: strconv.FormatInt(model.TaskExecuteErrorCodeSucceed, 10), Desc: "执行成功"},
			}
		}

		return w
	}
}
