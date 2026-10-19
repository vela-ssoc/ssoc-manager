package request

type Slug struct {
	Slug string `json:"slug" query:"slug" validate:"required"`
}
