package shipx

import "github.com/xgfone/ship/v5"

type RouteBinder interface {
	BindRoute(r *ship.RouteGroupBuilder) error
}

// RouteRegister 路由注册。
type RouteRegister interface {
	RegisterRoute(*ship.RouteGroupBuilder) error
}

type MetaRouter interface {
	Permission(c *ship.Context) (anonymous, logon, mfa bool)
}

type anonymous struct{}

func NewAnonymous() MetaRouter {
	return anonymous{}
}

func (a anonymous) Permission(*ship.Context) (bool, bool, bool) {
	return true, false, false
}
