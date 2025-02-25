package oauth

import (
	"context"
	"net/url"
)

type Configurer interface {
	// Configure 加载配置文件
	Configure(context.Context) (*Config, error)
}

type Config struct {
	URL          *url.URL
	ClientID     string
	ClientSecret string
	RedirectURL  string
}
