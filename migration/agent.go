package migration

import (
	"context"
	"log/slog"

	model2 "github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Agent struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewAgent(pdb polydb.Sourcer, log *slog.Logger) *Agent {
	return &Agent{
		pdb: pdb,
		log: log,
	}
}

func (agt *Agent) Exec(ctx context.Context) error {
	mdb := agt.pdb.MDB()
	coll := mdb.Agent()
	cnt, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		agt.log.Error("[agent] 查询新表出错", "err", err)
		return err
	} else if cnt != 0 {
		agt.log.Info("[agent] 数据无需再次迁移")
		return nil
	}

	bmap, err := agt.findBrokers(ctx)
	if err != nil {
		agt.log.Info("[agent] 查询中继节点出错", "err", err)
		return err
	}

	rdb := agt.pdb.RDB()
	tbl := rdb.Minion
	dao := tbl.WithContext(ctx)

	var lastID int64
	var offset int
	for {
		minions, err := dao.Where(tbl.ID.Gt(lastID)).Order(tbl.ID).
			Offset(offset).Limit(500).Find()
		if err != nil {
			agt.log.Error("[agent] 扫描旧数据出错", "err", err)
			return err
		}

		num := len(minions)
		if len(minions) == 0 {
			break
		}

		lastID = minions[num-1].ID
		if err = agt.saveAgent(ctx, bmap, minions); err != nil {
			agt.log.Error("[agent] 保存新数据出错", "err", err)
			return err
		}
	}

	return agt.moveTags(ctx)
}

func (agt *Agent) saveAgent(ctx context.Context, index map[int64]*model.Broker, minions model2.Minions) error {
	var dats []*model.Agent

	for _, m := range minions {
		v := agt.convertAgent(index, m)
		dats = append(dats, v)
	}

	mdb := agt.pdb.MDB()
	coll := mdb.Agent()
	_, err := coll.InsertMany(ctx, dats)

	return err
}

func (agt *Agent) convertAgent(index map[int64]*model.Broker, m *model2.Minion) *model.Agent {
	status := agt.convertAgentStatus(m.Status)

	data := &model.Agent{
		IDV1:      m.ID,
		MachineID: m.MachineID,
		Status:    status,
		Broker:    agt.lookupBroker(index, m.BrokerID),
		Unload:    m.Unload,
		TunnelStat: &model.TunnelStat{
			Inet: m.Inet,
		},
		ExecuteStat: &model.ExecuteStat{
			GOOS:       m.Goos,
			GOARCH:     m.Arch,
			Version:    m.Edition,
			Unstable:   m.Unstable,
			Customized: m.Customized,
		},
		HeartbeatAt: m.HeartbeatAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if status == model.AgentStatusOnline {
		data.DisconnectAt = m.HeartbeatAt
	}

	return data
}

func (agt *Agent) moveTags(ctx context.Context) error {
	rdb := agt.pdb.RDB()
	tbl := rdb.MinionTag
	dao := tbl.WithContext(ctx)
	dats, err := dao.Find()
	if err != nil {
		return err
	}

	tmap := agt.formatTags(dats)
	var mods []mongo.WriteModel
	for id, tags := range tmap {
		mod := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"idv1": id}).
			SetUpdate(bson.M{"$set": bson.M{"tags": tags}})
		mods = append(mods, mod)

		if len(mods) < 500 {
			continue
		}

		if err = agt.bulkWriteTags(ctx, mods); err != nil {
			return err
		}

		mods = mods[:0]
	}

	return agt.bulkWriteTags(ctx, mods)
}

func (agt *Agent) bulkWriteTags(ctx context.Context, mods []mongo.WriteModel) error {
	if len(mods) == 0 {
		return nil
	}

	mdb := agt.pdb.MDB()
	coll := mdb.Agent()
	opts := options.BulkWrite().SetOrdered(false)
	_, err := coll.BulkWrite(ctx, mods, opts)

	return err
}

func (agt *Agent) convertAgentStatus(s model2.MinionStatus) model.AgentStatus {
	switch s {
	case model2.MSDelete:
		return model.AgentStatusDeleted
	case model2.MSOnline:
		return model.AgentStatusOnline
	default:
		return model.AgentStatusOffline
	}
}

func (agt *Agent) convertAgentTagType(s model2.TagKind) model.AgentTagType {
	switch s {
	case model2.TkLifelong:
		return model.AgentTagBuiltin
	case model2.TkManual:
		return model.AgentTagManual
	default:
		return model.AgentTagAgent
	}
}

func (agt *Agent) findBrokers(ctx context.Context) (map[int64]*model.Broker, error) {
	mdb := agt.pdb.MDB()
	coll := mdb.Broker()
	dats, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	index := make(map[int64]*model.Broker, 8)
	for _, dat := range dats {
		index[dat.IDV1] = dat
	}

	return index, nil
}

func (agt *Agent) lookupBroker(index map[int64]*model.Broker, id int64) *model.BrokerEndpoint {
	brk := index[id]
	if brk == nil {
		return nil
	}

	return &model.BrokerEndpoint{
		ID:   brk.ID,
		IDV1: brk.IDV1,
		Slug: brk.Slug,
		Name: brk.Name,
	}
}

func (agt *Agent) formatTags(dats []*model2.MinionTag) map[int64]model.AgentTags {
	ret := make(map[int64]model.AgentTags, 4096)
	for _, dat := range dats {
		ret[dat.MinionID] = append(ret[dat.MinionID],
			model.AgentTag{
				Type: agt.convertAgentTagType(dat.Kind),
				Name: dat.Tag,
			},
		)
	}

	return ret
}
