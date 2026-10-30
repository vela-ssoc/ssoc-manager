package restapi

import (
	"net/http"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
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
		Data(shipx.NewAnonymous()).GET(oat.provider)
	g.ExposeV3.Group.Route("/oauth/dong").
		Data(shipx.NewAnonymous()).POST(oat.dong)

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

	op := model.Operator{
		ID:       user.ID,
		Company:  user.Company,
		Username: user.Username,
		Nickname: user.Nickname,
	}
	c.Any = op

	ret := &response.AuthInfo{
		ID:        user.ID,
		Company:   user.Company,
		Username:  user.Username,
		Nickname:  user.Nickname,
		Token:     user.Session.Token,
		AvatarURL: user.AvatarURL,
	}

	return c.JSON(http.StatusOK, ret)
}
