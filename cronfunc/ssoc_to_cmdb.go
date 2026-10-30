package cronfunc

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/vela-ssoc/ssoc-common-mb/crontab"
	model2 "github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SSOCToCMDB struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewSSOCToCMDB(pdb polydb.Sourcer, log *slog.Logger) *SSOCToCMDB {
	return &SSOCToCMDB{
		pdb: pdb,
		log: log,
	}
}

func (*SSOCToCMDB) Options() crontab.JobOptions {
	return crontab.JobOptions{
		ID:      "394846bc83b8aa4a970dd5d38c609fb8",
		Name:    "将 SSOC Agent 简要信息同步至 CMDB 表中",
		Sched:   cron.Every(time.Hour),
		Timeout: time.Hour,
	}
}

func (stc *SSOCToCMDB) Execute(ctx context.Context) error {
	now := time.Now()
	rdb, mdb := stc.pdb.RDB(), stc.pdb.MDB()

	tbl := rdb.Minion
	dao := tbl.WithContext(ctx)
	coll := mdb.CMDB()

	if cnt, _ := coll.CountDocuments(ctx, bson.D{}); cnt == 0 {
		stc.log.Warn("CMDB 表中还没有任何数据")
		return nil
	}

	var minionID int64
	const limit = 100

	for {
		minions, err := dao.Where(tbl.ID.Gt(minionID)).Limit(limit).Find()
		if err != nil {
			stc.log.Error("扫描探针表出错", "err", err)
			continue
		} else if len(minions) == 0 {
			break
		}

		minionID = minions[len(minions)-1].ID
		if err = stc.bulkWrite(ctx, minions, now); err != nil {
			stc.log.Error("更新 CMDB 中的 SSOC Agent 简要信息出错", "err", err)
		}
	}

	{
		filter := bson.M{"updated_at": bson.M{"$lte": now.Add(-time.Hour)}}
		update := bson.M{"$unset": bson.M{"ssoc_agent": true}}
		if _, err := coll.UpdateMany(ctx, filter, update); err != nil {
			return err
		}
	}

	return nil
}

func (stc *SSOCToCMDB) bulkWrite(ctx context.Context, minions model2.Minions, now time.Time) error {
	mods := make([]mongo.WriteModel, 0, len(minions))
	for _, m := range minions {
		simp := &model.SSOCAgentSimple{
			ID:         m.ID,
			MachineID:  m.MachineID,
			Version:    m.Edition,
			BrokerSlug: m.BrokerSlug,
			BrokerName: m.BrokerName,
		}
		switch m.Status {
		case model2.MSOnline:
			simp.Status = 1
		case model2.MSDelete:
			simp.Status = 2
		default:
		}

		mod := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"private_ip": m.Inet}).
			SetUpdate(bson.M{"$set": bson.M{"ssoc_agent": simp, "updated_at": now}})
		mods = append(mods, mod)
	}

	mdb := stc.pdb.MDB()
	coll := mdb.CMDB()
	opts := options.BulkWrite().SetOrdered(false)
	_, err := coll.BulkWrite(ctx, mods, opts)

	return err

}
