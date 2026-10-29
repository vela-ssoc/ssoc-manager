package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-manager/app/session"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/xgfone/ship/v5"
)

type OAuth struct {
	svc *service.OAuth
}

func NewOAuth(svc *service.OAuth) *OAuth {
	return &OAuth{svc: svc}
}

func (oat *OAuth) RegisterHTTP(g shipx.RouteGroup3) error {
	g.ExposeV3.Group.Route("/oauth/provider").
		GET(oat.provider)
	g.ExposeV3.Group.Route("/oauth/dong").
		POST(oat.dong)

	return nil
}

func (oat *OAuth) provider(c *ship.Context) error {
	var param request.OAuthProvider
	if err := c.BindQuery(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	ret, _ := oat.svc.Provider(ctx, param.Origin)

	return c.JSON(http.StatusOK, ret)
}

func (oat *OAuth) dong(c *ship.Context) error {
	var param request.OAuthDongLogin
	if err := c.Bind(&param); err != nil {
		return err
	}

	ctx := c.Request().Context()
	user, err := oat.svc.Dong(ctx, param.Code, param.RedirectURI)
	if err != nil {
		return err
	}

	cu := session.Issued(user)
	c.Any = cu
	if err = c.SetSession(cu.Token, cu); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, cu)
}
