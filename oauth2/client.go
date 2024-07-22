package oauth2

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/url"
	"strconv"

	"github.com/vela-ssoc/vela-manager/httpx"
)

type Client interface {
	Connect(ctx context.Context) (*QrcodeStatus, error)
	Status(ctx context.Context, uuid string, micro int64) (*QrcodeStatus, error)
	Exchange(ctx context.Context, code string) (*Userinfo, error)
}

func NewClient(cfg Configurer, cli httpx.Client) Client {
	return &dongClient{
		cfg: cfg,
		cli: cli,
	}
}

type dongClient struct {
	cfg Configurer
	cli httpx.Client
}

func (dc *dongClient) Connect(ctx context.Context) (*QrcodeStatus, error) {
	cfg, err := dc.cfg.Configure(ctx)
	if err != nil {
		return nil, err
	}

	body := &connectRequest{
		ClientID:     cfg.ClientID,
		RedirectURI:  cfg.RedirectURL,
		Scope:        "openid",
		ResponseType: "code",
	}

	addr := cfg.appendURL("/dongdong-auth/oauth2/connect")
	status := new(qrcodeStatus)
	result := &dongResponse{Data: status}
	if err = dc.cli.PostJSON(ctx, addr, nil, body, result); err != nil {
		return nil, err
	}
	if err = result.AutoError(); err != nil {
		return nil, err
	}

	return status.QrcodeStatus(), nil
}

func (dc *dongClient) Status(ctx context.Context, uuid string, micro int64) (*QrcodeStatus, error) {
	cfg, err := dc.cfg.Configure(ctx)
	if err != nil {
		return nil, err
	}

	stamp := strconv.FormatInt(micro, 10)
	quires := url.Values{"uuid": []string{uuid}, "t": []string{stamp}}
	addr := cfg.appendURL("/dongdong-auth/oauth2/connect/status") + "?" + quires.Encode()

	status := new(qrcodeStatus)
	result := &dongResponse{Data: status}
	if err = dc.cli.JSON(ctx, addr, nil, result); err != nil {
		return nil, err
	}
	if err = result.AutoError(); err != nil {
		return nil, err
	}

	return status.QrcodeStatus(), nil
}

func (dc *dongClient) Exchange(ctx context.Context, code string) (*Userinfo, error) {
	cfg, err := dc.cfg.Configure(ctx)
	if err != nil {
		return nil, err
	}

	data := url.Values{
		"grant_type":   []string{"authorization_code"},
		"redirect_uri": []string{cfg.RedirectURL},
		"code":         []string{code},
	}

	tokenAddr := cfg.appendURL("/dongdong-auth/oauth2/token")
	auth := cfg.ClientID + ":" + cfg.ClientSecret
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	header := http.Header{"Authorization": []string{basic}}

	ret := new(tokenResponse)
	if err = dc.cli.PostForm(ctx, tokenAddr, header, data, ret); err != nil {
		return nil, err
	}

	header.Set("Authorization", ret.Token())
	infoAddr := cfg.appendURL("/dongdong-auth/userinfo")
	info := new(dongUserinfoResponse)
	if err = dc.cli.JSON(ctx, infoAddr, header, &info); err != nil {
		return nil, err
	}

	return info.Userinfo(), nil
}
