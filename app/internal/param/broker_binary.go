package param

type BrokerDownload struct {
	IntID
	BrokerID int64 `query:"broker_id" validate:"required"`
}
