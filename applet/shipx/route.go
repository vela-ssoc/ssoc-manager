package shipx

import "github.com/xgfone/ship/v5"

type Router interface {
	Route(r *ship.RouteGroupBuilder) error
}

func BindRouters(rgb *ship.RouteGroupBuilder, rts []Router) error {
	for _, rt := range rts {
		if err := rt.Route(rgb); err != nil {
			return err
		}
	}
	return nil
}
