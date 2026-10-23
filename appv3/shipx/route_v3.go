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
