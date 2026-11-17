package agtool

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type AggregateInput struct {
	Collection string           `json:"collection" validate:"required"       jsonschema:"required" jsonschema_description:"集合名，必须在允许访问的集合列表中"`
	Pipeline   []map[string]any `json:"pipeline"   validate:"required,min=1" jsonschema:"required" jsonschema_description:"聚合管道数组，使用标准 MongoDB 聚合语法"`
}

type AggregateOutput struct {
	Collection string   `json:"collection"`
	Count      int      `json:"count"`
	Documents  []bson.M `json:"documents,omitempty"`
}

func (mgo *mongoInstance) aggregate(ctx context.Context, in AggregateInput) (*AggregateOutput, error) {
	if err := mgo.sess.Valid(in); err != nil {
		return nil, err
	}
	if err := mgo.isAllowedCollection(in.Collection); err != nil {
		return nil, err
	}

	coll := mgo.newCollection(in.Collection)
	opts := options.Aggregate().SetAllowDiskUse(true).SetMaxAwaitTime(mgo.timeout)

	cursor, err := coll.Aggregate(ctx, in.Pipeline, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []bson.M
	if err = cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	return &AggregateOutput{
		Collection: in.Collection,
		Count:      len(docs),
		Documents:  docs,
	}, nil
}
