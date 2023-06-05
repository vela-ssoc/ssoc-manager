package param

type LoginBase struct {
	Username string `json:"username" validate:"required,lte=20"`
	Password string `json:"password" validate:"required,gte=6,lte=32"`
}

type LoginSubmit struct {
	LoginBase
	HanID string `json:"code" validate:"required"`
	Dong  string `json:"dong" validate:"omitempty,len=6,numeric"`
}
