package migration

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	model2 "github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Broker struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewBroker(pdb polydb.Sourcer, log *slog.Logger) *Broker {
	return &Broker{
		pdb: pdb,
		log: log,
	}
}

func (brk *Broker) Exec(ctx context.Context) error {
	mdb := brk.pdb.MDB()
	coll := mdb.Broker()
	cnt, err := coll.CountDocuments(ctx, bson.D{})
	if err != nil {
		brk.log.Error("[broker] 查询新表出错", "err", err)
		return err
	} else if cnt != 0 {
		brk.log.Info("[broker] 数据无需再次迁移")
		return nil
	}

	rdb := brk.pdb.RDB()
	tbl := rdb.Broker
	dao := tbl.WithContext(ctx)

	olds, err := dao.Find()
	if err != nil || len(olds) == 0 {
		return err
	}

	brks := make([]*model.Broker, 0, len(olds))
	for _, old := range olds {
		dat := brk.convertBroker(old)
		brks = append(brks, dat)
	}

	for i, b := range olds {
		before, _, found := strings.Cut(b.Servername, ".")
		if found {
			brks[i].Slug = before
		} else {
			brks[i].Slug = strconv.FormatInt(b.ID, 32)
		}
	}
	_, err = coll.InsertMany(ctx, brks)

	return err
}

func (brk *Broker) convertBroker(old *model2.Broker) *model.Broker {
	return &model.Broker{
		IDV1:    old.ID,
		Name:    old.Name,
		Secret:  old.Secret,
		Exposes: brk.convertExposes(old.Servername, append(old.LAN, old.VIP...)),
		Config: model.BrokerConfig{
			Server: &model.BrokerServerConfig{Addr: old.Bind},
		},
		Status:      old.Status,
		HeartbeatAt: old.HeartbeatAt,
		CreatedAt:   old.CreatedAt,
		UpdatedAt:   old.UpdatedAt,
	}
}

func (brk *Broker) convertExposes(sn string, addrs []string) model.ExposeAddresses {
	dats := make(model.ExposeAddresses, 0, len(addrs))
	for i, addr := range addrs {
		exp := &model.ExposeAddress{
			Name:       "地址-" + strconv.Itoa(i),
			Addr:       addr,
			ServerName: sn,
		}
		dats = append(dats, exp)
	}

	return dats
}
