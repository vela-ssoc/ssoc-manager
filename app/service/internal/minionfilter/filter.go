package minionfilter

import (
	"context"
	"strconv"

	"github.com/vela-ssoc/ssoc-common-mb/dal/dyncond"
	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-common-mb/param/response"
	"github.com/vela-ssoc/ssoc-manager/param/mresponse"
	"gorm.io/gen"
	"gorm.io/gen/field"
)

func New(qry *query.Query) (*Filter, error) {
	flt := &Filter{qry: qry}

	opts := dyncond.Options{WhereCallback: flt.whereFilter, OrderCallback: flt.orderFilter}
	mods := []any{model.Minion{}, model.MinionTag{}}
	tbl, err := dyncond.ParseModels(qry, mods, opts)
	if err != nil {
		return nil, err
	}
	flt.tbl = tbl

	return flt, nil
}

type Filter struct {
	qry *query.Query
	tbl *dyncond.Tables
}

func (flt *Filter) Cond() *response.Cond {
	return response.ParseCond(flt.tbl)
}

func (flt *Filter) Delete(ctx context.Context, args *request.PageKeywordConditions) error {
	fdata := &request.KeywordConditions{Keywords: args.Keywords, Conditions: args.Conditions}
	err := flt.FindInBatches(ctx, fdata, nil, 100, func(tx gen.Dao, buf []*model.Minion) error {
		return nil
	})

	return err
}

func (flt *Filter) Page(ctx context.Context, args *request.PageKeywordConditions) (*response.Pages[*mresponse.MinionItem], error) {
	fdata := &request.KeywordConditions{Keywords: args.Keywords, Conditions: args.Conditions}
	wheres, err := flt.Wheres(fdata, nil)
	if err != nil {
		return nil, err
	}

	minion, minionTag := flt.qry.Minion, flt.qry.MinionTag
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
	sysInfo := flt.qry.SysInfo
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

func (flt *Filter) FindInBatches(ctx context.Context, args *request.KeywordConditions, excludes []string, batchSize int, cb func(tx gen.Dao, buf []*model.Minion) error) error {
	cond := &request.KeywordConditions{Keywords: args.Keywords, Conditions: args.Conditions}
	wheres, err := flt.Wheres(cond, excludes)
	if err != nil {
		return err
	}

	minion, minionTag := flt.qry.Minion, flt.qry.MinionTag
	minionTagDo := minionTag.WithContext(ctx)

	var buf []*model.Minion
	return minion.WithContext(ctx).
		LeftJoin(minionTagDo, minion.ID.EqCol(minionTag.MinionID)).
		Where(wheres...).
		FindInBatches(&buf, batchSize, func(tx gen.Dao, _ int) error {
			return cb(tx, buf)
		})
}

// Wheres 构造搜索条件。
func (flt *Filter) Wheres(args *request.KeywordConditions, excludes []string) ([]gen.Condition, error) {
	wheres, exprs, err := flt.tbl.CompileWhere(args.CondWhereInputs.Inputs(), false)
	if err != nil {
		return nil, err
	}

	minion, minionTag := flt.qry.Minion, flt.qry.MinionTag
	if len(excludes) != 0 {
		wheres = append(wheres, minion.Inet.NotIn(excludes...))
	}

	optionalLikes := []field.String{
		minionTag.Tag, minion.Inet, minion.Goos, minion.Arch, minion.Edition,
		minion.BrokerName, minion.Customized, minion.OrgPath, minion.Identity,
		minion.Category, minion.Comment, minion.IBu, minion.IDC, minion.OSRelease,
	}
	var likeFields []field.String
	for _, like := range optionalLikes {
		var exists bool
		for _, expr := range exprs {
			exists = flt.tbl.EqualsExpr(like, expr)
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

func (flt *Filter) compileWhere(args *request.KeywordConditions) ([]gen.Condition, error) {
	wheres, exprs, err := flt.tbl.CompileWhere(args.CondWhereInputs.Inputs(), false)
	if err != nil {
		return nil, err
	}

	minion, minionTag := flt.qry.Minion, flt.qry.MinionTag
	optionalLikes := []field.String{
		minionTag.Tag, minion.Inet, minion.Goos, minion.Arch, minion.Edition,
		minion.BrokerName, minion.Customized, minion.OrgPath, minion.Identity,
		minion.Category, minion.Comment, minion.IBu, minion.IDC,
	}

	var likeFields []field.String
	for _, like := range optionalLikes {
		var exists bool
		for _, expr := range exprs {
			exists = flt.tbl.EqualsExpr(like, expr)
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

func (flt *Filter) whereFilter(tbl *dyncond.Tables, w *dyncond.Where) *dyncond.Where {
	minion, minionTag := flt.qry.Minion, flt.qry.MinionTag
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

func (flt *Filter) orderFilter(_ *dyncond.Tables, _ *dyncond.Order) *dyncond.Order {
	return nil
}
