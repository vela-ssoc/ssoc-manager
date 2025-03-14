package casauth

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/url"

	"github.com/vela-ssoc/ssoc-common-mb/httpx"
)

type Configurer interface {
	Configure(ctx context.Context) (*url.URL, error)
}

type Client interface {
	Auth(ctx context.Context, name, passwd string) error
}

func NewClient(cfg Configurer, cli httpx.Client, log *slog.Logger) Client {
	return casClient{
		cfg: cfg,
		cli: cli,
		log: log,
	}
}

type casClient struct {
	cfg Configurer
	cli httpx.Client
	log *slog.Logger
}

func (c casClient) Auth(ctx context.Context, name, passwd string) error {
	attrs := []any{slog.String("name", name)}
	c.log.DebugContext(ctx, "开始CAS认证", attrs...)
	reqURL, err := c.cfg.Configure(ctx)
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		c.log.ErrorContext(ctx, "获取CAS配置错误", attrs...)
		return err
	}

	sum := md5.Sum([]byte(passwd))
	pwd := hex.EncodeToString(sum[:])
	query := reqURL.Query()
	query.Set("usrNme", name)
	query.Set("passwd", pwd)
	if query.Get("devTyp") == "" {
		query.Set("devTyp", "pc")
	}
	reqURL.RawQuery = query.Encode()
	strURL := reqURL.String()
	res := new(responseBody)
	if err = c.cli.JSON(ctx, strURL, nil, res); err != nil {
		attrs = append(attrs, slog.Any("error", err))
		c.log.ErrorContext(ctx, "请求CAS服务器错误", attrs...)
		return err
	}

	if err = res.checkError(); err != nil {
		attrs = append(attrs, slog.Any("error", err))
		c.log.WarnContext(ctx, "CAS认证失败", attrs...)
	} else {
		c.log.InfoContext(ctx, "CAS认证通过", attrs...)
	}

	return err
}

var errorCodes = map[string]string{
	"01": "密码错误",
	"03": "用户不存在",
	"04": "设备类型错误",
	"09": "用户名或密码为空",
}

// reply sso 认证服务的响应报文
type responseBody struct {
	RspCde string `json:"rspCde"` // 业务响应码
	RspMsg string `json:"rspMsg"` // 响应消息
}

func (rb responseBody) checkError() error {
	code := rb.RspCde
	if code == "00" {
		return nil
	}

	msg := rb.RspMsg
	if msg == "" {
		msg = errorCodes[code]
	}
	if msg == "" {
		msg = "认证错误（CAS-code: " + code + "）"
	}

	return errors.New(msg)
}
