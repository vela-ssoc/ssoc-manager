package param

type TagUpdate struct {
	ID   int64    `json:"id,string" validate:"required"`
	Tags []string `json:"tags"      validate:"gt=0,unique,dive,tag"`
}
