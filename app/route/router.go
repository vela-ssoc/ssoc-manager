package route

import "github.com/xgfone/ship/v5"

// Router 路由绑定接口
type Router interface {
	// Route 绑定路由
	// anon: 无需登录即可访问
	// bearer: 从 Header 中获取 Token
	// basic: 支持 bearer 认证方式的同时还支持 basic auth 认证
	Route(anon, bearer, basic *ship.RouteGroupBuilder)
}
