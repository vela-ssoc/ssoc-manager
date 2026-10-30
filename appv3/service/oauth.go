package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-manager/appv3/response"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/oauth2"
)

type OAuth struct {
	pdb polydb.Sourcer
	log *slog.Logger
}

func NewOAuth(pdb polydb.Sourcer, log *slog.Logger) *OAuth {
	return &OAuth{pdb: pdb, log: log}
}

func (oat *OAuth) Provider(ctx context.Context, origin string) (*response.OAuthProvider, error) {
	cfg, err := oat.currentProvider(ctx)
	if err != nil {
		return nil, err
	}
	// 匹配最符合的 redirectURI
	ret := &response.OAuthProvider{ClientID: cfg.ClientID, AuthorizationEndpoint: cfg.AuthorizationEndpoint}
	if num := len(cfg.RedirectURIs); origin == "" || num <= 1 {
		if num == 1 {
			ret.RedirectURI = cfg.RedirectURIs[0]
		}
		return ret, nil
	}

	// 错误的 origin 就返回第一个 RedirectURI
	ori, err := url.Parse(origin)
	if err != nil {
		ret.RedirectURI = cfg.RedirectURIs[0]
		return ret, nil
	}
	for _, ru := range cfg.RedirectURIs {
		uri, err := url.Parse(ru)
		if err != nil {
			continue
		}

		if uri.Scheme == ori.Scheme && uri.Host == ori.Host {
			ret.RedirectURI = ru
			break
		}
	}
	if ret.RedirectURI == "" {
		ret.RedirectURI = cfg.RedirectURIs[0]
	}

	return ret, nil
}

func (oat *OAuth) Dong(ctx context.Context, code, redirectURI string) (*model.User, error) {
	now := time.Now()
	info, err := oat.exchangeDong(ctx, code, redirectURI)
	if err != nil {
		oat.log.Warn("咚咚 OAuth 认证失败", "err", err)
		return nil, err
	}

	mdb := oat.pdb.MDB()
	coll := mdb.User()
	filter := bson.M{"username": info.SUB, "company": info.CompanyCode}
	user, err := coll.FindOne(ctx, filter)
	if err != nil {
		oat.log.Warn("用户尚未注册本系统", "userinfo", info)
		return nil, errcode.ErrAuthFailed
	}
	if !user.OAuth {
		oat.log.Warn("用户不可使用 OAuth 认证", "userinfo", info)
		return nil, errcode.ErrAuthFailed
	}
	if !user.Enabled {
		oat.log.Warn("用户已禁用")
		return nil, errcode.ErrAuthFailed
	}
	if !user.ExpiresAt.IsZero() && user.ExpiresAt.Before(now) {
		oat.log.Warn("用户已过期")
		return nil, errcode.ErrAuthFailed
	}

	buf := make([]byte, 50)
	_, _ = rand.Read(buf)
	token := "session." + hex.EncodeToString(buf)

	sess := model.UserSession{Token: token, IssuedAt: now, LastedAt: now}
	update := bson.M{"$set": bson.M{"avatar_url": info.Picture, "session": sess}}
	if _, err = coll.UpdateByID(ctx, user.ID, update); err != nil {
		return nil, err
	}

	user.AvatarURL = info.Picture
	user.Session = sess

	return user, nil
}

func (oat *OAuth) currentProvider(ctx context.Context) (*model.OAuthConfig, error) {
	mdb := oat.pdb.MDB()
	coll := mdb.OAuthConfig()
	cfg, err := coll.Get(ctx)
	if err != nil {
		oat.log.Info("获取 OAuth 提供商信息错误", "err", err)
		return nil, errcode.ErrForbiddenOAuth
	} else if !cfg.Enabled {
		return nil, errcode.ErrForbiddenOAuth
	}

	return cfg, nil
}

func (oat *OAuth) exchangeDong(ctx context.Context, code, redirectURI string) (*model.OAuthDongUserinfo, error) {
	prov, err := oat.currentProvider(ctx)
	if err != nil {
		return nil, err
	}

	cfg := &oauth2.Config{
		ClientID:     prov.ClientID,
		ClientSecret: prov.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  prov.AuthorizationEndpoint,
			TokenURL: prov.TokenEndpoint,
		},
		RedirectURL: redirectURI,
		Scopes:      prov.Scopes,
	}
	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		oat.log.ErrorContext(ctx, "咚咚认证密钥交换出错", "err", err)
		return nil, err
	}
	cli := cfg.Client(ctx, tok)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, prov.UserinfoEndpoint, nil)
	if err != nil {
		return nil, err
	}

	res, err := cli.Do(req)
	if err != nil {
		oat.log.ErrorContext(ctx, "咚咚认证请求用户信息出错", "err", err)
		return nil, err
	}
	defer res.Body.Close()

	info := new(model.OAuthDongUserinfo)
	if err = json.NewDecoder(res.Body).Decode(info); err != nil {
		return nil, err
	}

	return info, nil
}
