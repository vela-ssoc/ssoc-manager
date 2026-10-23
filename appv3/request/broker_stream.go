package request

type BrokerSubstreamKill struct {
	Slug
	ID uint64 `json:"id,string" query:"id" validate:"required"`
}
