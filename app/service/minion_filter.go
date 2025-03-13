package service

import (
	"context"
	"strconv"

	"github.com/vela-ssoc/ssoc-manager/app/service/internal/minionfilter"
	"github.com/vela-ssoc/ssoc-manager/param/mresponse"
	"github.com/vela-ssoc/vela-common-mb/dal/dyncond"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/param/request"
	"github.com/vela-ssoc/vela-common-mb/param/response"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

func NewMinionFilter(qry *query.Query) (*minionfilter.Filter, error) {
	return minionfilter.New(qry)
}

type MinionFilter struct {
	qry *query.Query
	tbl *dyncond.Tables
}

func (mf *MinionFilter) Cond() *response.Cond {
	return response.ParseCond(mf.tbl)
}

func (mf *MinionFilter) Page(ctx context.Context, args *request.PageKeywordConditions) (*response.Pages[*mresponse.MinionItem], error) {
	wheres, err := mf.compileWhere(&request.KeywordConditions{Keywords: args.Keywords, Conditions: args.Conditions})
	if err != nil {
		return nil, err
	}

	minion, minionTag := mf.qry.Minion, mf.qry.MinionTag
	minionTagDo := minionTag.WithContext(ctx)

	pages := response.NewPages[*mresponse.MinionItem](args.PageSize())
	minionDo := minion.WithContext(ctx).
		Distinct(minion.ID).
		LeftJoin(minionTagDo, minion.ID.EqCol(minionTag.MinionID)).
		Where(wheres...)
	cnt, err := minionDo.Count()
	if err != nil {
		return nil, err
	} else if cnt == 0 {
		return pages.Empty(), nil
	}

	var minionIDs []int64
	if err = minionDo.Scopes(pages.FP(cnt)).
		Order(minion.ID).
		Scan(&minionIDs); err != nil {
		return nil, err
	}

	var records []*mresponse.MinionItem
	if err = minion.WithContext(ctx).
		Where(minion.ID.In(minionIDs...)).
		Scan(&records); err != nil {
		return nil, err
	}

	minionTags, _ := minionTagDo.Where(minionTag.MinionID.In(minionIDs...)).Find()
	sysInfo := mf.qry.SysInfo
	sysInfos, _ := sysInfo.WithContext(ctx).Where(sysInfo.ID.In(minionIDs...)).Find()

	index := make(map[int64]*mresponse.MinionItem, 16)
	for _, record := range records {
		index[record.ID] = record
	}
	for _, info := range sysInfos {
		if item := index[info.ID]; item != nil {
			item.MemFree = info.MemFree
			item.MemTotal = info.MemTotal
			item.CPUCore = info.CPUCore
		}
	}
	for _, tag := range minionTags {
		if item := index[tag.MinionID]; item != nil {
			item.Tags = append(item.Tags, tag.Tag)
		}
	}

	return pages.SetRecords(records), nil
}

func (mf *MinionFilter) compileWhere(args *request.KeywordConditions) ([]gen.Condition, error) {
	wheres, exprs, err := mf.tbl.CompileWhere(args.CondWhereInputs.Inputs(), false)
	if err != nil {
		return nil, err
	}

	minion, minionTag := mf.qry.Minion, mf.qry.MinionTag
	optionalLikes := []field.String{
		minionTag.Tag, minion.Inet, minion.Goos, minion.Arch, minion.Edition,
		minion.BrokerName, minion.Customized, minion.OrgPath, minion.Identity,
		minion.Category, minion.Comment, minion.IBu, minion.IDC,
	}

	var likeFields []field.String
	for _, like := range optionalLikes {
		var exists bool
		for _, expr := range exprs {
			exists = mf.tbl.EqualsExpr(like, expr)
		}
		if !exists {
			likeFields = append(likeFields, like)
		}
	}

	likes := args.Likes(likeFields...)
	if len(likes) != 0 {
		wheres = append(wheres, field.Or(likes...))
	}

	return wheres, nil
}

func (mf *MinionFilter) whereFilter(tbl *dyncond.Tables, w *dyncond.Where) *dyncond.Where {
	minion, minionTag := mf.qry.Minion, mf.qry.MinionTag
	ignores := []field.Expr{
		minionTag.ID, minionTag.MinionID, minionTag.Kind,
		minion.CreatedAt, minion.UpdatedAt,
	}
	for _, ignore := range ignores {
		if tbl.EqualsExpr(w.Expr, ignore) {
			return nil
		}
	}

	if tbl.EqualsExpr(w.Expr, minionTag.Tag) {
		w.Operators = []dyncond.Operator{dyncond.Eq, dyncond.In}
	} else if tbl.EqualsExpr(w.Expr, minion.Goos) {
		w.Operators = []dyncond.Operator{dyncond.Eq, dyncond.Neq, dyncond.In, dyncond.NotIn}
		w.Enums = dyncond.Enums{
			{Key: "linux", Desc: "Linux"},
			{Key: "windows", Desc: "Windows"},
			{Key: "darwin", Desc: "MacOS"},
		}
	} else if tbl.EqualsExpr(w.Expr, minion.Arch) {
		w.Operators = []dyncond.Operator{dyncond.Eq, dyncond.Neq, dyncond.In, dyncond.NotIn}
		w.Enums = dyncond.Enums{
			{Key: "amd64", Desc: "amd64"},
			{Key: "386", Desc: "386"},
			{Key: "arm64", Desc: "arm64"},
			{Key: "arm32", Desc: "arm32"},
		}
	} else if tbl.EqualsExpr(w.Expr, minion.Status) {
		w.Enums = dyncond.Enums{
			{Key: strconv.FormatInt(int64(model.MSOnline), 10), Desc: "在线"},
			{Key: strconv.FormatInt(int64(model.MSOffline), 10), Desc: "离线"},
			{Key: strconv.FormatInt(int64(model.MSDelete), 10), Desc: "删除"},
		}
	}

	return w
}

func (mf *MinionFilter) orderFilter(_ *dyncond.Tables, _ *dyncond.Order) *dyncond.Order {
	return nil
}
