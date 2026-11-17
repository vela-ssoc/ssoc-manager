package agtool

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type FindOneInput struct {
	Collection string         `json:"collection"          validate:"required" jsonschema:"required" jsonschema_description:"集合名，必须在允许访问的集合列表中"`
	Filter     map[string]any `json:"filter,omitempty"                          jsonschema_description:"MongoDB 查询条件，例如 {\"_id\":\"...\"} 或 {\"machine_id\":\"...\"}"`
	Projection map[string]any `json:"projection,omitempty"                      jsonschema_description:"返回字段投影；不传返回全部字段"`
	Sort       map[string]any `json:"sort,omitempty"                            jsonschema_description:"排序，例如 {\"heartbeat_at\":-1}"`
}

type FindOneOutput struct {
	Collection string `json:"collection"`
	Found      bool   `json:"found"`
	Document   bson.M `json:"document,omitempty"`
}

func (mgo *mongoInstance) findOne(ctx context.Context, in FindOneInput) (*FindOneOutput, error) {
	if err := mgo.sess.Valid(in); err != nil {
		return nil, err
	}
	if err := mgo.isAllowedCollection(in.Collection); err != nil {
		return nil, err
	}

	coll := mgo.pdb.MDB().Database().Collection(in.Collection)
	opts := options.FindOne()
	if len(in.Sort) > 0 {
		opts.SetSort(in.Sort)
	}
	if len(in.Projection) > 0 {
		opts.SetProjection(in.Projection)
	}

	var doc bson.M
	if err := coll.FindOne(ctx, in.Filter, opts).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return &FindOneOutput{Collection: in.Collection}, nil
		}
		return nil, err
	}

	return &FindOneOutput{
		Collection: in.Collection,
		Found:      true,
		Document:   doc,
	}, nil
}
