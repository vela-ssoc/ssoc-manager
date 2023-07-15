package param

type ViewHTML struct {
	IntID
	Secret string `json:"secret" query:"secret" validate:"required,lte=255"`
}
