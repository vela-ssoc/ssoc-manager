package request

type Pages struct {
	Page int64 `json:"page" query:"page" form:"page" validate:"gte=0"`
	Size int64 `json:"size" query:"size" form:"size" validate:"gte=0,lte=1000"`
}

type Keywords struct {
	Keyword string `json:"keyword" query:"keyword" form:"keyword"`
}

type PageKeywords struct {
	Pages
	Keywords
}
