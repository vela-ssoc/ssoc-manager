package shipx

import "github.com/xgfone/ship/v5"

type RouteBinder interface {
	BindRoute(r *ship.RouteGroupBuilder) error
}

// RouteRegister 路由注册。
type RouteRegister interface {
	RegisterRoute(*ship.RouteGroupBuilder) error
}
