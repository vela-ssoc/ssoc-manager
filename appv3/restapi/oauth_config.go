package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type OAuthConfig struct {
	svc *service.OAuthConfig
}

func NewOAuthConfig(svc *service.OAuthConfig) *OAuthConfig {
	return &OAuthConfig{svc: svc}
}

func (oat *OAuthConfig) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/oauth/config").
		GET(oat.get).
		POST(oat.upsert)

	return nil
}

func (oat *OAuthConfig) get(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, _ := oat.svc.Get(ctx)

	return c.JSON(http.StatusOK, ret)
}

func (oat *OAuthConfig) upsert(c *ship.Context) error {
	var param request.OAuthConfigUpsert
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return oat.svc.Upsert(ctx, param)
}
