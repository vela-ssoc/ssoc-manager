package param

type ThirdCustomizedCreate struct {
	Name   string `json:"name"   validate:"required,lte=10"`
	Icon   string `json:"icon"   validate:"required,lte=65500"`
	Remark string `json:"remark" validate:"lte=1000"`
}

type ThirdCustomizedUpdate struct {
	IntID
	Icon   string `json:"icon"   validate:"required,lte=65500"`
	Remark string `json:"remark" validate:"lte=1000"`
}
