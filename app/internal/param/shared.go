package param

import "encoding/json"

type SharedBucketKey struct {
	Bucket string `json:"bucket" query:"bucket" validate:"lte=200"`
	Key    string `json:"key"    query:"key"    validate:"lte=200"`
}

type SharedAuditPage struct {
	Page
	Bucket string `json:"bucket" query:"bucket" validate:"required,lte=200"`
	Key    string `json:"key"    query:"key"    validate:"required,lte=200"`
}

type SharedUpdate struct {
	Bucket string          `json:"bucket" validate:"required,lte=200"`
	Key    string          `json:"key"    validate:"required,lte=200"`
	Value  json.RawMessage `json:"value"  validate:"lte=65535"`
}
