package service

import (
	"context"
	"io"
	"log/slog"

	"github.com/vela-ssoc/vela-common-mb/dal/condition"
	"github.com/vela-ssoc/vela-common-mb/dal/gridfs2"
	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/pagination"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/applet/manager/request"
	"github.com/vela-ssoc/vela-manager/applet/manager/response"
	"gorm.io/gen/field"
)

func NewGrid(gfs gridfs.FS, qry *query.Query, log *slog.Logger) *Grid {
	mod := new(model.GridFile)
	ctx := context.Background()
	tbl := qry.GridFile
	db := tbl.WithContext(ctx).UnderlyingDB()
	ignores := []field.Expr{tbl.MIME, tbl.MD5, tbl.SHA1, tbl.SHA256, tbl.Shard}
	opt := &condition.ParserOptions{IgnoreOrder: ignores, IgnoreWhere: ignores}
	cond, _ := condition.ParseModel(db, mod, opt)

	return &Grid{
		gfs:  gfs,
		qry:  qry,
		log:  log,
		cond: cond,
	}
}

type Grid struct {
	gfs  gridfs.FS
	qry  *query.Query
	log  *slog.Logger
	cond *condition.Cond
}

func (gd *Grid) Cond() *response.Cond {
	return response.ReadCond(gd.cond)
}

func (gd *Grid) Open(ctx context.Context, id int64) (gridfs.File, error) {
	return gd.gfs.Open(ctx, id)
}

func (gd *Grid) Create(ctx context.Context, name string, r io.Reader) error {
	_, err := gd.gfs.Create(ctx, name, r)
	return err
}

func (gd *Grid) Page(ctx context.Context, req *request.PageCondition) (*pagination.Result[*model.GridFile], error) {
	tbl := gd.qry.GridFile
	input := req.AllInputs()
	wheres := gd.cond.CompileWheres(input.Where)
	orders := gd.cond.CompileOrders(input.Order)
	dao := tbl.WithContext(ctx).Where(wheres...)
	cnt, err := dao.Count()
	if err != nil {
		return nil, err
	}
	pager := pagination.NewPager[*model.GridFile](req.PageSize())
	if cnt == 0 {
		empty := pager.Empty()
		return empty, nil
	}

	dats, err := dao.Order(orders...).Scopes(pager.Scope(cnt)).Find()
	if err != nil {
		return nil, err
	}
	ret := pager.Result(dats)

	return ret, nil
}
