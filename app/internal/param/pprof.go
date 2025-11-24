package param

type PprofLoad struct {
	Node   string `query:"node"   validate:"required"`
	Second int    `query:"second" validate:"gte=0,lte=60"`
}

type PprofDump struct {
	ID     int64  `json:"id,string" query:"id"     validate:"required,lte=100"`
	Type   string `json:"type"      query:"type"   validate:"oneof=heap profile"`
	Second int    `json:"second"    query:"second" validate:"gte=0,lte=300"`
}
