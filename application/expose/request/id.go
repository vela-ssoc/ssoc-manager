package request

type Int64ID struct {
	ID int64 `json:"id,string" form:"id" query:"id" validate:"required"`
}
