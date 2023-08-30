package param

type PprofLoad struct {
	Node   string `query:"node"   validate:"required"`
	Second int    `query:"second" validate:"gte=0,lte=60"`
}
