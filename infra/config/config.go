package config

import (
	"context"
	"strings"

	"github.com/vela-ssoc/vela-common-mb/dbms"
	"github.com/vela-ssoc/vela-manager/oauth2"
)

// Config 配置参数
type Config struct {
	Server   Server      `json:"server"   yaml:"server"`   // HTTP 服务配置
	Database dbms.Config `json:"database" yaml:"database"` // 数据库配置
	Logger   Logger      `json:"logger"   yaml:"logger"`   // 日志配置
	Section  Section     `json:"section"  yaml:"section"`  // 其他信息
	Oauth    Oauth       `json:"oauth"    yaml:"oauth"`
	SIEM     SIEM        `json:"siem"     yaml:"siem"`
}

type SIEM struct {
	URL   string `json:"url"   yaml:"url"`
	Token string `json:"token" yaml:"token"`
}

type Oauth struct {
	URL          string `json:"url"           yaml:"url"`
	ClientID     string `json:"client_id"     yaml:"client_id"`
	ClientSecret string `json:"client_secret" yaml:"client_secret"`
	RedirectURL  string `json:"redirect_url"  yaml:"redirect_url"`
}

func (o Oauth) Configure(_ context.Context) (*oauth2.Config, error) {
	rawURL := strings.TrimRight(o.URL, "/")

	return &oauth2.Config{
		URL:          rawURL,
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		RedirectURL:  o.RedirectURL,
	}, nil
}
