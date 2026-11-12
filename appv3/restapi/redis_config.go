package restapi

import (
	"net/http"
	"strings"

	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/appv3/service"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/xgfone/ship/v5"
)

type RedisConfig struct {
	svc *service.RedisConfig
}

func NewRedisConfig(svc *service.RedisConfig) *RedisConfig {
	return &RedisConfig{svc: svc}
}

func (rc *RedisConfig) RegisterHTTP(g shipx.RouteGroup) error {
	g.ExposeV3.Group.Route("/redis-config").
		GET(rc.get).    // 获取 Redis 配置
		POST(rc.upsert) // 修改/新增 Redis 配置
	g.ExposeV3.Group.Route("/redis-config/safe-commands").GET(rc.safeCommands)
	g.ExposeV3.Group.Route("/redis/info").GET(rc.info)
	g.ExposeV3.Group.Route("/redis/keys").GET(rc.keys)
	g.ExposeV3.Group.Route("/redis/value").GET(rc.value)
	g.ExposeV3.Group.Route("/redis/do").POST(rc.do)

	return nil
}

func (rc *RedisConfig) get(c *ship.Context) error {
	ctx := c.Request().Context()
	ret, _ := rc.svc.Get(ctx)

	return c.JSON(http.StatusOK, ret)
}

func (rc *RedisConfig) upsert(c *ship.Context) error {
	var param request.RedisConfigUpsert
	if err := c.Bind(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()

	return rc.svc.Upsert(ctx, param)
}

func (rc *RedisConfig) safeCommands(c *ship.Context) error {
	cmds := rc.svc.SafeCommands()
	dats := make([]string, 0, len(cmds))
	for k := range cmds {
		dats = append(dats, k)
	}
	ret := response.NewData(dats)

	return c.JSON(http.StatusOK, ret)
}

func (rc *RedisConfig) do(c *ship.Context) error {
	var param request.RedisDo
	if err := c.Bind(&param); err != nil {
		return err
	}

	if !param.Privilege {
		cmd := strings.TrimSpace(param.Args[0])
		cmd = strings.ToUpper(cmd)
		safes := rc.svc.SafeCommands()
		if _, exists := safes[cmd]; !exists {
			return errcode.ErrForbidden
		}
	}

	ctx := c.Request().Context()
	dat, err := rc.svc.Do(ctx, param.AnyArgs())
	if err != nil {
		return ship.ErrBadRequest.New(err)
	}
	ret := response.NewData(dat)

	return c.JSON(http.StatusOK, ret)
}

func (rc *RedisConfig) info(c *ship.Context) error {
	var param request.RedisInfo
	if err := c.BindQuery(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()
	dat, err := rc.svc.Info(ctx, param.Section)
	if err != nil {
		return err
	}
	ret := response.NewData(dat)

	return c.JSON(http.StatusOK, ret)
}

func (rc *RedisConfig) keys(c *ship.Context) error {
	var param request.RedisKeys
	if err := c.BindQuery(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()
	ret, err := rc.svc.Keys(ctx, param.Cursor, param.Match, param.SN())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}

func (rc *RedisConfig) value(c *ship.Context) error {
	var param request.RedisValue
	if err := c.BindQuery(&param); err != nil {
		return err
	}
	ctx := c.Request().Context()
	ret, err := rc.svc.Value(ctx, param.Key, param.Cursor, param.Match, param.SN())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, ret)
}
