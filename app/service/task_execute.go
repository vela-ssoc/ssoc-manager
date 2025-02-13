package service

import (
	"context"
	"log/slog"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/param/request"
	"github.com/vela-ssoc/vela-manager/param/response"
	"gorm.io/gen"
)

type TaskExecute struct {
	qry *query.Query
	log *slog.Logger
}

func NewTaskExecute(qry *query.Query) *TaskExecute {
	return &TaskExecute{
		qry: qry,
	}
}

func (tex *TaskExecute) Page(ctx context.Context, args *request.TaskExecutePage) (*response.Pages[*model.TaskExecute], error) {
	tbl := tex.qry.TaskExecute
	var wheres []gen.Condition
	if taskID := args.TaskID; taskID > 0 {
		wheres = append(wheres, tbl.TaskID.Eq(taskID))
	}
	dao := tbl.WithContext(ctx).Where(wheres...).
		Scopes(args.Regexps(tbl.Name, tbl.Intro))

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

func (tex *TaskExecute) Items(ctx context.Context, args *request.TaskExecuteItems) (*response.Pages[*model.TaskExecuteItem], error) {
	tbl := tex.qry.TaskExecuteItem
	dao := tbl.WithContext(ctx).Where(tbl.TaskID.Eq(args.TaskID), tbl.ExecID.Eq(args.ExecID))
	page := response.NewPages[*model.TaskExecuteItem](args.Page, args.Size)
	cnt, err := dao.Count()
	if err != nil {
		return nil, err
	} else if cnt == 0 {
		return page.Empty(), nil
	}

	records, err := dao.Scopes(page.FP(cnt)).Find()
	if err != nil {
		return nil, err
	}

	return page.SetRecords(records), nil
}
