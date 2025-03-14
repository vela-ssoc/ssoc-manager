package oauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/vela-ssoc/ssoc-common-mb/httpx"
)

type Client interface {
	Connect(ctx context.Context) (*QrcodeStatus, error)
	Status(ctx context.Context, uuid string, micro int64) (*QrcodeStatus, error)
	Exchange(ctx context.Context, code string) (*Userinfo, error)
}

func NewClient(cfg Configurer, cli httpx.Client, log *slog.Logger) Client {
	return &dongClient{
		cfg: cfg,
		cli: cli,
		log: log,
	}
}

type dongClient struct {
	cfg Configurer
	cli httpx.Client
	log *slog.Logger
}

func (dc *dongClient) Connect(ctx context.Context) (*QrcodeStatus, error) {
	attrs := []any{slog.String("step", "connect")}
	dc.log.DebugContext(ctx, "OAUTH", attrs...)
	cfg, err := dc.cfg.Configure(ctx)
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		dc.log.ErrorContext(ctx, "获取OAUTH配置错误", attrs...)
		return nil, err
	}

	body := &connectRequest{
		ClientID:     cfg.ClientID,
		RedirectURI:  cfg.RedirectURL,
		Scope:        "openid",
		ResponseType: "code",
	}

	strURL := cfg.URL.JoinPath("/oauth2/connect").String()
	status := new(qrcodeStatus)
	result := &dongResponse{Data: status}
	if err = dc.cli.PostJSON(ctx, strURL, nil, body, result); err != nil {
		attrs = append(attrs, slog.Any("error", err))
		dc.log.ErrorContext(ctx, "OAUTH", attrs...)
		return nil, err
	}
	if err = result.AutoError(); err != nil {
		attrs = append(attrs, slog.Any("error", err))
		dc.log.ErrorContext(ctx, "OAUTH服务器响应错误", attrs...)
		return nil, err
	}

	return status.QrcodeStatus(), nil
}

func (dc *dongClient) Status(ctx context.Context, uuid string, micro int64) (*QrcodeStatus, error) {
	attrs := []any{slog.String("step", "connect")}
	dc.log.DebugContext(ctx, "OAUTH", attrs...)
	cfg, err := dc.cfg.Configure(ctx)
	if err != nil {
		dc.log.ErrorContext(ctx, "获取OAUTH配置错误", slog.Any("error", err))
		return nil, err
	}

	stamp := strconv.FormatInt(micro, 10)
	reqURL := cfg.URL.JoinPath("/oauth2/connect/status")
	query := reqURL.Query()
	query.Set("uuid", uuid)
	query.Set("t", stamp)
	reqURL.RawQuery = query.Encode()
	strURL := reqURL.String()

	status := new(qrcodeStatus)
	result := &dongResponse{Data: status}
	if err = dc.cli.JSON(ctx, strURL, nil, result); err != nil {
		attrs = append(attrs, slog.Any("error", err))
		dc.log.ErrorContext(ctx, "OAUTH", attrs...)
		return nil, err
	}
	if err = result.AutoError(); err != nil {
		attrs = append(attrs, slog.Any("error", err))
		dc.log.ErrorContext(ctx, "OAUTH服务器响应错误", attrs...)
		return nil, err
	}

	return status.QrcodeStatus(), nil
}

func (dc *dongClient) Exchange(ctx context.Context, code string) (*Userinfo, error) {
	attrs := []any{slog.String("step", "connect")}
	dc.log.DebugContext(ctx, "OAUTH", attrs...)
	cfg, err := dc.cfg.Configure(ctx)
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		dc.log.ErrorContext(ctx, "获取OAUTH配置错误", attrs...)
		return nil, err
	}

	data := url.Values{
		"grant_type":   []string{"authorization_code"},
		"redirect_uri": []string{cfg.RedirectURL},
		"code":         []string{code},
	}

	tokenURL := cfg.URL.JoinPath("/oauth2/token").String()
	auth := cfg.ClientID + ":" + cfg.ClientSecret
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	header := http.Header{"Authorization": []string{basic}}

	ret := new(tokenResponse)
	if err = dc.cli.PostForm(ctx, tokenURL, header, data, ret); err != nil {
		return nil, err
	}
	if info := dc.parseJWT(ret.IDToken); info != nil { // fast path
		dc.log.DebugContext(ctx, "通过解析JWT获取用户信息")
		return info, nil
	}

	header.Set("Authorization", ret.Token())
	infoURL := cfg.URL.JoinPath("/userinfo").String()
	info := new(dongUserinfoResponse)
	dc.log.InfoContext(ctx, "请求 /userinfo 获取用户信息")
	if err = dc.cli.JSON(ctx, infoURL, header, info); err != nil {
		attrs = append(attrs, slog.Any("error", err))
		dc.log.ErrorContext(ctx, "OAUTH服务器响应错误", attrs...)
		return nil, err
	}
	userinfo := info.Userinfo()
	attrs = append(attrs, slog.String("name", userinfo.Name), slog.String("job_number", userinfo.SUB))
	dc.log.InfoContext(ctx, "用户信息获取成功", attrs...)

	return userinfo, nil
}

func (dc *dongClient) parseJWT(token string) *Userinfo {
	sn := strings.SplitN(token, ".", 3)
	if len(sn) != 3 {
		return nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(sn[1])
	if err != nil {
		return nil
	}
	info := new(dongUserinfoResponse)
	if err = json.Unmarshal(raw, info); err != nil {
		return nil
	}

	return info.Userinfo()
}
