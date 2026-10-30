package cronfunc

import (
	"context"
	"strings"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-manager/integration/ywcmdb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (cj *CMDB) dutiesConvert(duties ywcmdb.Duties) model.SimpleDuties {
	ret := make(model.SimpleDuties, 0, len(duties))
	for _, d := range duties {
		sd := model.SimpleDuty{JobNumber: d.EmployeeID, Name: d.Nickname, Company: d.UType}
		ret = append(ret, sd)
	}

	return ret
}

func (cj *CMDB) newCMDBBulkWriteModel(data *model.CMDB, now time.Time) mongo.WriteModel {
	data.UpdatedAt = now
	filter := bson.D{bson.E{Key: "ci_id", Value: data.CIID}}
	update := bson.M{"$set": data, "$setOnInsert": bson.M{"created_at": now}}

	return mongo.NewUpdateOneModel().
		SetUpsert(true).
		SetFilter(filter).
		SetUpdate(update)
}

func (cj *CMDB) bulkWriteCMDB(ctx context.Context, mods []mongo.WriteModel) error {
	if len(mods) == 0 {
		return nil
	}

	coll := cj.pdb.MDB().CMDB()
	opts := options.BulkWrite().SetOrdered(false)
	_, err := coll.BulkWrite(ctx, mods, opts)

	return err
}

// splitPrivateIPs CMDB2 PrivateIP 字段格式不统一，有的用 / 分隔，有的用 , ，中英文逗号分隔。
func (*CMDB) splitPrivateIPs(strs []string) []string {
	var temp []string
	for _, pip := range strs {
		var splits []string
		if strings.Contains(pip, ",") {
			splits = strings.Split(pip, ",")
		} else if strings.Contains(pip, "/") {
			splits = strings.Split(pip, "/")
		} else if strings.Contains(pip, "，") {
			splits = strings.Split(pip, "，")
		} else {
			splits = append(splits, pip)
		}
		temp = append(temp, splits...)
	}

	return temp
}
