package shipx

import "github.com/xgfone/ship/v5"

type HTTPRegistrar interface {
	RegisterHTTP(g RouteGroup3) error
}

type RouteEndpoint struct {
	Path  string                  // 路径前缀，例如 /api/v3
	Group *ship.RouteGroupBuilder // 路由注册分组
}

type RouteGroup3 struct {
	ExposeV1 RouteEndpoint
	ExposeV3 RouteEndpoint
	BrokerV1 RouteEndpoint
	BrokerV3 RouteEndpoint
}

func (g RouteGroup3) RegisterRoutes(hs []HTTPRegistrar) error {
	for _, h := range hs {
		if err := h.RegisterHTTP(g); err != nil {
			return err
		}
	}

	return nil
}

func (g RouteGroup3) RegisterV1Routes(v1s []V1HTTPRegistrar) {
	grp := g.ExposeV1.Group
	for _, v1 := range v1s {
		v1.Route(grp, grp, grp)
	}
}

type V1HTTPRegistrar interface {
	// Route 绑定路由
	// anon: 无需登录即可访问
	// bearer: 从 Header 中获取 Token
	// basic: 支持 bearer 认证方式的同时还支持 basic auth 认证
	Route(anon, bearer, basic *ship.RouteGroupBuilder)
}
