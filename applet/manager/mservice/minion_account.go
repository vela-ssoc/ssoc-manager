package mservice

import (
	"context"
	"log/slog"

	"github.com/vela-ssoc/vela-common-mb/dal/model"
	"github.com/vela-ssoc/vela-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mb/param/response"
	"github.com/vela-ssoc/vela-manager/applet/manager/mrequest"
	"gorm.io/gen"
)

type MinionAccount struct {
	qry *query.Query
	log *slog.Logger
}

func NewMinionAccount(qry *query.Query, log *slog.Logger) *MinionAccount {
	return &MinionAccount{
		qry: qry,
		log: log,
	}
}

func (ma *MinionAccount) Page(ctx context.Context, args *mrequest.MinionAccountPage) (*response.Pages[*model.MinionAccount], error) {
	pages := response.NewPages[*model.MinionAccount](args.PageSize())
	tbl := ma.qry.MinionAccount
	var wheres []gen.Condition
	if mid := args.MinionID; mid > 0 {
		wheres = append(wheres, tbl.MinionID.Eq(mid))
	}
	if name := args.Name; name != "" {
		wheres = append(wheres, tbl.Name.Regexp(name))
	}
	tblDo := tbl.WithContext(ctx).Where(wheres...)
	cnt, err := tblDo.Count()
	if err != nil {
		return nil, err
	} else if cnt == 0 {
		return pages.Empty(), nil
	}

	dats, err := tblDo.Scopes(pages.FP(cnt)).Find()
	if err != nil {
		return nil, err
	}

	return pages.SetRecords(dats), nil
}
