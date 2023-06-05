package param

type StoreUpsert struct {
	ID      string `json:"id"    validate:"required,lte=50"`
	Value   []byte `json:"value" validate:"required,gte=1,lte=65535"`
	Desc    string `json:"desc"  validate:"lte=255"`
	Version int64  `json:"version"`
}
