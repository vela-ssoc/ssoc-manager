package agtool

import (
	"context"
)

type CountInput struct {
	Collection string         `json:"collection"          validate:"required" jsonschema:"required" jsonschema_description:"集合名，必须在允许访问的集合列表中"`
	Filter     map[string]any `json:"filter,omitempty"                          jsonschema_description:"MongoDB 查询条件，例如 {\"status\":1}"`
}

type CountOutput struct {
	Collection string `json:"collection"`
	Count      int64  `json:"count"`
}

func (mgo *mongoInstance) countDocuments(ctx context.Context, in CountInput) (*CountOutput, error) {
	if err := mgo.sess.Valid(in); err != nil {
		return nil, err
	}
	if err := mgo.isAllowedCollection(in.Collection); err != nil {
		return nil, err
	}

	coll := mgo.newCollection(in.Collection)
	count, err := coll.CountDocuments(ctx, in.Filter)
	if err != nil {
		return nil, err
	}

	return &CountOutput{
		Collection: in.Collection,
		Count:      count,
	}, nil
}
