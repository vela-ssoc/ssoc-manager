package linkhub

type Handler interface {
	TaskSync(bid, mid int64)
}
