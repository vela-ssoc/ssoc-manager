package cronfunc

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/vela-ssoc/ssoc-common-mb/crontab"
	model2 "github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CMDBToSSOC struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewCMDBToSSOC(pdb polydb.Sourcer, log *slog.Logger) *CMDBToSSOC {
	return &CMDBToSSOC{
		pdb: pdb,
		log: log,
	}
}

func (cts *CMDBToSSOC) Options() crontab.JobOptions {
	return crontab.JobOptions{
		ID:      "8c6083464a19fb83948bc4b8aa3dd5d3",
		Name:    "将 CMDB 简要信息同步至 SSOC Agent 表中",
		Sched:   cron.Every(27 * time.Hour),
		Timeout: time.Hour,
	}
}

func (cts *CMDBToSSOC) Execute(ctx context.Context) error {
	coll := cts.pdb.MDB().CMDB()
	if cnt, _ := coll.CountDocuments(ctx, bson.D{}); cnt == 0 {
		cts.log.Warn("CMDB 表中还没有任何数据")
		return nil
	}

	tbl := cts.pdb.RDB().Minion
	dao := tbl.WithContext(ctx)

	var minionID int64
	const limit = 100

	for {
		minions, err := dao.Where(tbl.ID.Gt(minionID)).Limit(limit).Find()
		if err != nil {
			cts.log.Error("扫描探针表出错", "err", err)
			continue
		} else if len(minions) == 0 {
			break
		}

		minionID = minions[len(minions)-1].ID

		cts.sync(ctx, minions)
	}

	return nil
}

func (cts *CMDBToSSOC) sync(ctx context.Context, minions model2.Minions) {
	tbl := cts.pdb.RDB().Minion
	dao := tbl.WithContext(ctx)
	coll := cts.pdb.MDB().CMDB()
	for _, mon := range minions {
		cdb, err := coll.FindOne(ctx, bson.M{"private_ip": mon.Inet})
		if err != nil {
			continue
		}

		_, _ = dao.Where(tbl.ID.Eq(mon.ID)).
			UpdateColumnSimple(
				tbl.OrgPath.Value(cdb.Department),
				tbl.Identity.Value(cdb.BaoleijiIdentity),
				tbl.Comment.Value(cdb.Comment),
				tbl.IBu.Value(cdb.Business),
				tbl.IDC.Value(cdb.IDC),
				tbl.OSRelease.Value(cdb.OSVersion),
			)
	}
}
