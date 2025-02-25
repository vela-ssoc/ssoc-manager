package mrequest

import "github.com/vela-ssoc/vela-common-mb/param/request"

type MinionAccountPage struct {
	request.PageKeywords
	MinionID int64 `json:"minion_id,string" query:"minion_id"`

	// Deprecated:
	Name string `json:"name"             query:"name"`
}
