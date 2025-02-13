package request

type Int64ID struct {
	ID int64 `json:"id,string" query:"id" form:"id" validate:"required,gt=0"`
}

type Int64IDOptional struct {
	ID int64 `json:"id,string" query:"id" form:"id"`
}
