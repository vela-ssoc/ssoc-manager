package mrequest

type ExecID struct {
	ExecID int64 `json:"exec_id,string" query:"exec_id" form:"exec_id" validate:"required,gt=0"`
}
