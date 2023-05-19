package middle

import (
	"net/http"
	"strings"

	"github.com/vela-ssoc/vela-manager/errcode"
	"github.com/xgfone/ship/v5"
)

type Authenticate interface {
	Bearer(ship.Handler) ship.Handler
	Basic(ship.Handler) ship.Handler
}

func Auth(headerKey, queryKey string) Authenticate {
	return &authMiddle{
		headerKey: headerKey,
		queryKey:  queryKey,
	}
}

type authMiddle struct {
	headerKey string // Header 中 Token 的 key
	queryKey  string // Query 参数中 Token 的 key
}

func (am *authMiddle) Bearer(h ship.Handler) ship.Handler {
	return func(c *ship.Context) error {
		token := am.bearer(c.Request())
		cu, err := c.GetSession(token)
		if err != nil {
			return errcode.ErrUnauthorized
		}
		c.Any = cu

		return h(c)
	}
}

func (am *authMiddle) Basic(h ship.Handler) ship.Handler {
	return func(c *ship.Context) error {
		token := am.basic(c.Request())
		cu, err := c.GetSession(token)
		if err != nil {
			c.SetRespHeader(ship.HeaderWWWAuthenticate, "Basic realm=\"Restricted\"")
			return errcode.ErrUnauthorized
		}
		c.Any = cu

		return h(c)
	}
}

// bearer 从 Header 或 Query 参数中获取 Token，
// 在 Header 中获取的 token 可以有 Bearer 前缀，也可以没有。
func (am *authMiddle) bearer(req *http.Request) (token string) {
	const prefix = "Bearer "
	if token = req.Header.Get(am.headerKey); token == "" {
		token = req.URL.Query().Get(am.queryKey)
	} else {
		token = strings.TrimPrefix(token, prefix)
	}
	return
}

// basic 先从 Header 或 Query 参数中获取 Token，再从 BasicAuth 中获取。
func (am *authMiddle) basic(req *http.Request) (token string) {
	if token = am.bearer(req); token == "" || am.headerKey == ship.HeaderAuthorization {
		if _, passwd, ok := req.BasicAuth(); ok {
			token = passwd
		}
	}
	return
}
