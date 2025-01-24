package service

import (
	"context"
	"strings"

	"github.com/vela-ssoc/vela-common-mb/cmdb2"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-manager/app/internal/mapstruct"
	"gorm.io/gen/field"
	"gorm.io/gorm/clause"
)

type Cmdb2Service interface {
	Rsync(ctx context.Context) error
}

func Cmdb2(qry *query.Query, cli cmdb2.Client) Cmdb2Service {
	return &cmdb2Service{
		qry: qry,
		cli: cli,
	}
}

type cmdb2Service struct {
	qry *query.Query
	cli cmdb2.Client
}

func (biz *cmdb2Service) Rsync(ctx context.Context) error {
	offset, limit := 0, 100
	for {
		inets, err := biz.scroll(ctx, offset, limit)
		if err != nil || len(inets) == 0 {
			return err
		}
		offset += limit
		hms, err := biz.fetchCmdb2(ctx, inets)
		if err != nil || len(hms) == 0 {
			continue
		}
		biz.updateCmdb2(ctx, inets, hms)
	}
}

func (biz *cmdb2Service) scroll(ctx context.Context, offset int, limit int) ([]string, error) {
	tbl := biz.qry.Minion
	ret := make([]string, 0, limit)
	err := tbl.WithContext(ctx).
		Distinct(tbl.Inet).
		Offset(offset).
		Limit(limit).
		Scan(&ret)

	return ret, err
}

func (biz *cmdb2Service) fetchCmdb2(ctx context.Context, inets []string) (map[string]*cmdb2.Server, error) {
	length := len(inets)
	srvs, err := biz.cli.Servers(ctx, inets, 0, length)
	if err != nil {
		return nil, err
	}
	hms := make(map[string]*cmdb2.Server, length)
	for _, srv := range srvs {
		for _, ip := range srv.PrivateIP {
			hms[ip] = srv
		}
	}

	return hms, nil
}

func (biz *cmdb2Service) updateCmdb2(ctx context.Context, inets []string, hms map[string]*cmdb2.Server) {
	for _, inet := range inets {
		srv := hms[inet]
		if srv == nil {
			continue
		}

		{
			ops := make([]string, 0, 10)
			for _, m := range srv.OpDutyMain {
				ops = append(ops, m.Nickname)
			}

			tbl := biz.qry.Minion
			dao := tbl.WithContext(ctx)
			args := []field.AssignExpr{
				tbl.OrgPath.Value(srv.Department),
				tbl.Identity.Value(srv.BaoleijiIdentity),
				tbl.Category.Value(srv.AppName),
				tbl.OpDuty.Value(strings.Join(ops, ",")),
				tbl.Comment.Value(srv.Comment),
				tbl.IDC.Value(srv.IDC),
			}
			_, _ = dao.Where(tbl.Inet.Eq(inet)).UpdateSimple(args...)
		}
		{
			ass := mapstruct.Cmdb2Server(srv)
			ass.Inet = inet
			tbl := biz.qry.Cmdb2
			_ = tbl.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Save(ass)
		}
	}
}
