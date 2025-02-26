package mrequest

import "github.com/vela-ssoc/vela-common-mb/param/request"

type TaskExecutePage struct {
	request.Pages
	request.Keywords
	TaskID int64 `json:"task_id,string" query:"task_id" form:"task_id"`
}

type TaskExecuteItems struct {
	request.Pages
	request.Keywords
	TaskID int64 `json:"task_id,string" query:"task_id" form:"task_id" validate:"required,gt=0"`
	ExecID int64 `json:"exec_id,string" query:"exec_id" form:"exec_id" validate:"required,gt=0"`
}
