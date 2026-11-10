package cronfunc

import (
	"context"
	"log/slog"
	"net"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/crontab"
	model2 "github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SSOCCMDBSync struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewSSOCCMDBSync(pdb polydb.Sourcer, log *slog.Logger) *SSOCCMDBSync {
	return &SSOCCMDBSync{
		pdb: pdb,
		log: log,
	}
}

func (*SSOCCMDBSync) Options() crontab.JobOptions {
	return crontab.JobOptions{
		ID:   "394846bc83b8aa380dd5dfb8c6094a97",
		Name: "SSOC 与 CMDB 信息互相同步",
		// Sched:   cron.Every(2 * time.Hour),
		Expr:    "28 9 * * *",
		Timeout: time.Hour,
	}
}

func (scs *SSOCCMDBSync) Execute(ctx context.Context) error {
	mdb := scs.pdb.MDB()
	coll := mdb.CMDB()

	// 如果 CMDB 没有任何数据，就没必要继续执行了。
	if cnt, _ := coll.CountDocuments(ctx, bson.D{}); cnt == 0 {
		scs.log.Warn("CMDB 数据为空")
		return nil
	}

	scs.syncAllBroker(ctx)
	scs.syncAllAgent(ctx)

	return nil
}

// syncAllBroker 中继节点。
func (scs *SSOCCMDBSync) syncAllBroker(ctx context.Context) {
	mdb := scs.pdb.MDB()
	bcoll := mdb.Broker()
	ccoll := mdb.CMDB()
	brks, err := bcoll.Find(ctx, bson.D{})
	if err != nil {
		scs.log.Error("查询中继节点信息出错", "err", err)
		return
	}
	num := len(brks)
	if num == 0 {
		scs.log.Info("没有查询到任何中继节点")
		return
	}

	bmods, cmods := make([]mongo.WriteModel, 0, num), make([]mongo.WriteModel, 0, num)
	for _, brk := range brks {
		cdb, _ := scs.findBrokerCMDB(ctx, brk)
		if cdb == nil {
			continue
		}

		nc := scs.convertNodeCMDB(cdb)
		bsimp := &model.CMDBSSOCBrokerSimple{
			ID:     brk.ID,
			Status: brk.Status,
			Slug:   brk.Slug,
			Name:   brk.Name,
		}
		if stat := brk.ExecuteStat; stat != nil {
			bsimp.Version = stat.Version
		}

		bmod := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": brk.ID}).
			SetUpdate(bson.M{"$set": bson.M{"cmdb": nc}})
		cmod := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": cdb.ID}).
			SetUpdate(bson.M{"$set": bson.M{"ssoc_broker": bsimp}})
		bmods = append(bmods, bmod)
		cmods = append(cmods, cmod)
	}

	opts := options.BulkWrite().SetOrdered(false)
	_, _ = bcoll.BulkWrite(ctx, bmods, opts)
	_, _ = ccoll.BulkWrite(ctx, cmods, opts)
}

// syncAllAgent 探针节点。
func (scs *SSOCCMDBSync) syncAllAgent(ctx context.Context) {
	tbl := scs.pdb.RDB().Minion
	dao := tbl.WithContext(ctx)

	var lastID int64
	for {
		mons, err := dao.Where(tbl.ID.Gt(lastID)).Limit(200).Find()
		if err != nil {
			scs.log.Error("扫描探针表出错", "err", err)
			break
		}
		num := len(mons)
		if num == 0 {
			break
		}

		lastID = mons[num-1].ID

		scs.syncAgents(ctx, mons)
	}
}

func (scs *SSOCCMDBSync) syncAgents(ctx context.Context, mons model2.Minions) {
	tbl := scs.pdb.RDB().Minion
	dao := tbl.WithContext(ctx)

	mods := make([]mongo.WriteModel, 0, len(mons))
	for _, mon := range mons {
		cdb, err := scs.findCMDB(ctx, mon.Inet)
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

		asimp := &model.CMDBSSOCAgentSimple{
			IDV1:      mon.ID,
			Status:    scs.convertAgentStatus(mon.Status),
			MachineID: mon.MachineID,
			Version:   mon.Edition,
			Broker: &model.BrokerEndpoint{
				IDV1: mon.BrokerID,
				Slug: mon.BrokerSlug,
				Name: mon.BrokerName,
			},
		}
		mod := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": cdb.ID}).
			SetUpdate(bson.M{"$set": bson.M{"ssoc_agent": asimp}})
		mods = append(mods, mod)
	}
	if len(mods) == 0 {
		return
	}

	mdb := scs.pdb.MDB()
	coll := mdb.CMDB()
	opts := options.BulkWrite().SetOrdered(false)
	_, _ = coll.BulkWrite(ctx, mods, opts)
}

func (scs *SSOCCMDBSync) findBrokerCMDB(ctx context.Context, brk *model.Broker) (*model.CMDB, string) {
	var ips []string
	for _, exp := range brk.Exposes {
		if host, _, err := net.SplitHostPort(exp.Addr); err == nil && host != "" {
			ips = append(ips, host)
		}
	}
	if ts := brk.TunnelStat; ts != nil && ts.Inet != "" {
		ips = append(ips, ts.Inet)
	}

	for _, ip := range ips {
		if v, _ := scs.findCMDB(ctx, ip); v != nil {
			return v, ip
		}
	}

	return nil, ""
}

func (scs *SSOCCMDBSync) findCMDB(ctx context.Context, inet string) (*model.CMDB, error) {
	coll := scs.pdb.MDB().CMDB()
	return coll.FindOne(ctx, bson.M{"private_ip": inet})
}

func (*SSOCCMDBSync) convertNodeCMDB(c *model.CMDB) *model.NodeCMDB {
	return &model.NodeCMDB{
		CIType:           c.CIType,
		AppDuty:          c.AppDuty,
		AppName:          c.AppName,
		BaoleijiIdentity: c.BaoleijiIdentity,
		Comment:          c.Comment,
		CPU:              c.CPU,
		CPUCount:         c.CPUCount,
		Department:       c.Department,
		Description:      c.Description,
		IDC:              c.IDC,
		IPv6:             c.IPv6,
		OSVersion:        c.OSVersion,
		OpDuty:           c.OpDuty,
		PrivateIP:        c.PrivateIP,
	}
}

func (*SSOCCMDBSync) convertAgentStatus(s model2.MinionStatus) model.AgentStatus {
	switch s {
	case model2.MSOnline:
		return model.AgentStatusOnline
	case model2.MSDelete:
		return model.AgentStatusDeleted
	default:
		return model.AgentStatusOffline
	}
}
