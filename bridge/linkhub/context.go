package linkhub

import "context"

type contextKey struct {
	name string
}

var brokerCtxKey = &contextKey{name: "broker-context"}

func FromContext(ctx context.Context) Peer {
	if ctx != nil {
		val, _ := ctx.Value(brokerCtxKey).(Peer)
		return val
	}
	return nil
}

func withContext(c *spdyServerConn) context.Context {
	return context.WithValue(context.Background(), brokerCtxKey, c)
}
