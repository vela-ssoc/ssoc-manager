package agtool

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FindInput struct {
	Collection string         `json:"collection"          validate:"required"        jsonschema:"required" jsonschema_description:"集合名，必须在允许访问的集合列表中"`
	Filter     map[string]any `json:"filter,omitempty"                                 jsonschema_description:"MongoDB 查询条件，例如 {\"status\":1}"`
	Projection map[string]any `json:"projection,omitempty"                             jsonschema_description:"返回字段投影，例如 {\"hostname\":1,\"status\":1}"`
	Sort       map[string]any `json:"sort,omitempty"                                   jsonschema_description:"排序，例如 {\"heartbeat_at\":-1}"`
	Limit      int64          `json:"limit"                validate:"required,min=1,max=200" jsonschema:"required" jsonschema_description:"返回条数，范围 1~200"`
	Skip       int64          `json:"skip,omitempty"       validate:"omitempty,min=0,max=10000" jsonschema_description:"跳过条数，最大 10000"`
}

type FindOutput struct {
	Collection string   `json:"collection"`
	Count      int64    `json:"count"`
	Documents  []bson.M `json:"documents,omitempty"`
}

func (mgo *mongoInstance) find(ctx context.Context, in FindInput) (*FindOutput, error) {
	if err := mgo.sess.Valid(in); err != nil {
		return nil, err
	}
	if err := mgo.isAllowedCollection(in.Collection); err != nil {
		return nil, err
	}

	coll := mgo.newCollection(in.Collection)
	opts := options.Find().
		SetLimit(in.Limit).
		SetSkip(in.Skip).
		SetMaxAwaitTime(mgo.timeout).
		SetAllowDiskUse(true)
	if len(in.Sort) > 0 {
		opts.SetSort(in.Sort)
	}
	if len(in.Projection) > 0 {
		opts.SetProjection(in.Projection)
	}

	cursor, err := coll.Find(ctx, in.Filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	return &FindOutput{
		Collection: in.Collection,
		Count:      int64(len(docs)),
		Documents:  docs,
	}, nil
}
