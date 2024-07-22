package oauth2

import "context"

type Configurer interface {
	// Configure 加载配置文件
	Configure(context.Context) (*Config, error)
}

type Config struct {
	URL          string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func (c Config) appendURL(path string) string {
	return c.URL + path
}
