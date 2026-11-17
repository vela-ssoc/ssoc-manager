package agtool

import "context"

type DistinctInput struct {
	Collection string         `json:"collection"          validate:"required" jsonschema:"required" jsonschema_description:"集合名，必须在允许访问的集合列表中"`
	Field      string         `json:"field"               validate:"required" jsonschema:"required" jsonschema_description:"要去重的字段名，支持点号路径，例如 broker.slug"`
	Filter     map[string]any `json:"filter,omitempty"                          jsonschema_description:"先过滤再取去重值，例如 {\"status\":1}"`
}

type DistinctOutput struct {
	Collection string `json:"collection"`
	Field      string `json:"field"`
	Count      int    `json:"count"`
	Values     []any  `json:"values,omitempty"`
}

func (mgo *mongoInstance) distinct(ctx context.Context, in DistinctInput) (*DistinctOutput, error) {
	if err := mgo.sess.Valid(in); err != nil {
		return nil, err
	}
	if err := mgo.isAllowedCollection(in.Collection); err != nil {
		return nil, err
	}

	coll := mgo.newCollection(in.Collection)
	result := coll.Distinct(ctx, in.Field, in.Filter)
	if err := result.Err(); err != nil {
		return nil, err
	}

	var values []any
	if err := result.Decode(&values); err != nil {
		return nil, err
	}

	return &DistinctOutput{
		Collection: in.Collection,
		Field:      in.Field,
		Count:      len(values),
		Values:     values,
	}, nil
}
