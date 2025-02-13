package request

type Pages struct {
	Page int64 `query:"page" json:"page" form:"page" validate:"gte=0"`
	Size int64 `query:"size" json:"size" form:"size" validate:"gte=0,lte=1000"`
}
