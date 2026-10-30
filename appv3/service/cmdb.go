package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/repository"
	"github.com/vela-ssoc/ssoc-common-mb/memcache"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewCMDB(pdb polydb.Sourcer, log *slog.Logger) *CMDB {
	cdb := &CMDB{
		pdb: pdb,
		log: log,
	}
	cdb.dept = memcache.NewTTL(time.Minute, cdb.department)

	return cdb
}

type CMDB struct {
	pdb  polydb.Sourcer
	log  *slog.Logger
	dept *memcache.TTL[model.CMDBNodes]
}

func (cdb *CMDB) Department(ctx context.Context) (model.CMDBNodes, error) {
	return cdb.dept.Load(ctx)
}

func (cdb *CMDB) Assets(ctx context.Context, param request.CMDBAssets) (*repository.Pages[model.CMDB], error) {
	ands := bson.A{bson.M{"ci_type": bson.M{"$ne": model.CIDepartment}}}
	if len(param.CIType) != 0 {
		ands = append(ands, bson.M{"ci_type": bson.M{"$in": param.CIType}})
	}

	fields := []string{
		"private_ip", "host_ip", "name", "comment", "os_version", "domain_name", "record_value", "vip",
		"rserver.rserver", "idc", "comment", "department", "baoleiji_identity", "hostname", "app_name",
	}
	if likes := param.Regex(fields...); len(likes) != 0 {
		ands = append(ands, bson.M{"$or": likes})
	}

	if dept := param.Department; dept != "" {
		prefix := "^" + dept + "/"
		regex := bson.Regex{Pattern: prefix}
		cond := bson.A{bson.M{"department": dept}, bson.M{"department": regex}}
		ands = append(ands, bson.M{"$or": cond})
	}

	filter := make(bson.M, 4)
	if len(ands) != 0 {
		filter["$and"] = ands
	}

	page, size := param.PSN()
	coll := cdb.pdb.MDB().CMDB()
	opts := options.Find().SetSort(bson.D{{Key: "ci_id", Value: -1}})

	return coll.Page(ctx, filter, page, size, opts)
}

func (cdb *CMDB) department(ctx context.Context) (model.CMDBNodes, error) {
	coll := cdb.pdb.MDB().CMDB()
	proj := bson.M{"ci_type": 1, "department": 1, "ci_id": 1, "name": 1}
	filter := bson.D{bson.E{Key: "ci_type", Value: model.CIDepartment}}
	opts := options.Find().SetSort(bson.D{{Key: "db_id", Value: -1}}).SetProjection(proj)
	dats, err := coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	ret := model.CMDBs(dats).Treeing()

	return ret, nil
}
