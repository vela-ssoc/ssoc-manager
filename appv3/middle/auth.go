package middle

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-manager/appv3/shipx"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/xgfone/ship/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type auth struct {
	pdb polydb.Sourcer
	exp time.Duration
	log *slog.Logger
}

func NewAuth(pdb polydb.Sourcer, exp time.Duration, log *slog.Logger) ship.Middleware {
	if exp < time.Minute {
		exp = time.Minute
	}

	ath := &auth{
		pdb: pdb,
		exp: exp,
		log: log,
	}

	return ath.handle
}

func (ath *auth) handle(h ship.Handler) ship.Handler {
	return func(c *ship.Context) error {
		mr, ok := c.Route.Data.(shipx.MetaRouter)
		if ok {
			anonymous, _, _ := mr.Permission(c)
			if anonymous {
				return h(c)
			}
		}

		token := c.GetReqHeader(ship.HeaderAuthorization)
		if token == "" && c.Method() == http.MethodGet {
			token = c.Query("authorization")
		}
		token, _ = strings.CutPrefix(token, "Bearer ")
		if !strings.HasPrefix(token, "session.") {
			return errcode.ErrUnauthorized
		}

		parent := c.Request().Context()
		coll := ath.pdb.MDB().User()
		filter := bson.M{"session.token": token}
		user, err := coll.FindOne(parent, filter)
		if err != nil {
			return errcode.ErrUnauthorized
		}

		// 检查用户
		now := time.Now()
		if !user.Enabled || // 用户被禁用
			(!user.ExpiresAt.IsZero() && user.ExpiresAt.Before(now)) || // 用户到期
			user.Session.IssuedAt.Add(ath.exp).Before(now) { // 令牌过期
			return errcode.ErrUnauthorized
		}

		// 刷新最后一次使用时间，存在并发刷新，但不影响功能，后续优化。
		if now.Sub(user.Session.LastedAt) > time.Minute {
			ctx, cancel := context.WithTimeout(parent, 3*time.Second)
			update := bson.M{"$set": bson.M{"session.lasted_at": now}}
			_, _ = coll.UpdateOne(ctx, filter, update)
			cancel()
		}

		return h(c)
	}
}
