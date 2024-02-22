package param

type SharedBucketKey struct {
	Bucket string `json:"bucket" query:"bucket" validate:"lte=200"`
	Key    string `json:"key"    query:"key"    validate:"lte=200"`
}
