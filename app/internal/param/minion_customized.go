package param

type MinionCustomizedCreate struct {
	Name string `json:"name" validate:"required,lte=10"`
	Icon string `json:"icon" validate:"required,lte=65500"`
}
