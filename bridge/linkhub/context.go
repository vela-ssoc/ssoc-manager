package linkhub

import "context"

type contextKey struct {
	name string
}

var brokerCtxKey = &contextKey{name: "broker-context"}

func Ctx(ctx context.Context) any {
	if ctx != nil {
		val, _ := ctx.Value(brokerCtxKey).(any)
		return val
	}
	return nil
}
