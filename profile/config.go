package profile

import (
	"context"
	"strings"
	"time"

	"github.com/vela-ssoc/vela-common-mb/cmdb2"
	"gorm.io/gorm/logger"
)

// Config 全局配置文件。
type Config struct {
	Active   string   `json:"active"   yaml:"active"`   // 多环境配置
	Server   Server   `json:"server"   yaml:"server"`   // HTTP 服务
	Database Database `json:"database" yaml:"database"` // 数据库
	Logger   Logger   `json:"logger"   yaml:"logger"`   // 日志
	Cmdb2    Cmdb2    `json:"cmdb2"    yaml:"cmdb2"`    // CMDB2
	Oauth    Oauth    `json:"oauth"    yaml:"oauth"`
}

type Server struct {
	Addr    string   `json:"addr"    yaml:"addr"`                            // 监听地址
	Cert    string   `json:"cert"    yaml:"cert"   validate:"lte=500"`       // 证书
	Pkey    string   `json:"pkey"    yaml:"pkey"   validate:"lte=500"`       // 私钥
	Static  string   `json:"static"  yaml:"static" validate:"lte=500"`       // 静态资源路径
	Vhosts  []string `json:"vhosts"  yaml:"vhosts" validate:"dive,required"` // 虚拟主机头
	Session duration `json:"session" yaml:"session"`
	CDN     string   `json:"cdn"     yaml:"cdn"`
}

type Database struct {
	DSN         string   `json:"dsn"           yaml:"dsn"           validate:"required,lte=255"`                      // 数据库连接
	Level       string   `json:"level"         yaml:"level"         validate:"omitempty,oneof=DEBUG INFO WARN ERROR"` // 日志输出级别
	Migrate     bool     `json:"migrate"       yaml:"migrate"`                                                        // 是否合并差异
	MaxOpenConn int      `json:"max_open_conn" yaml:"max_open_conn" validate:"gte=0"`                                 // 最大连接数
	MaxIdleConn int      `json:"max_idle_conn" yaml:"max_idle_conn" validate:"gte=0"`                                 // 最大空闲连接数
	MaxLifeTime duration `json:"max_life_time" yaml:"max_life_time" validate:"gte=0"`                                 // 连接最大存活时长
	MaxIdleTime duration `json:"max_idle_time" yaml:"max_idle_time" validate:"gte=0"`                                 // 空闲连接最大时长
}

type Oauth struct {
	CAS          string `json:"cas"           yaml:"cas"` // 旧版认证接口
	URL          string `json:"url"           yaml:"url"`
	ClientID     string `json:"client_id"     yaml:"client_id"`
	ClientSecret string `json:"client_secret" yaml:"client_secret"`
	RedirectURL  string `json:"redirect_url"  yaml:"redirect_url"`
}

type Cmdb2 struct {
	URL       string `json:"url"        yaml:"url"`
	AccessKey string `json:"access_key" yaml:"access_key"`
	SecretKey string `json:"secret_key" yaml:"secret_key"`
}

func (c Cmdb2) Configure(_ context.Context) (*cmdb2.Config, error) {
	cfg := &cmdb2.Config{
		URL:       c.URL,
		AccessKey: c.AccessKey,
		SecretKey: c.SecretKey,
	}

	return cfg, nil
}

func (c Config) GormLevel() logger.LogLevel {
	lvl := c.Database.Level
	if lvl == "" {
		lvl = c.Logger.Level
	}

	switch strings.ToUpper(lvl) {
	case "ERROR":
		return logger.Error
	case "WARN":
		return logger.Warn
	default:
		return logger.Info
	}
}

type duration time.Duration

func (du *duration) UnmarshalText(bs []byte) error {
	pd, err := time.ParseDuration(string(bs))
	if err == nil {
		*du = duration(pd)
	}

	return err
}

func (du *duration) String() string {
	return time.Duration(*du).String()
}

func (du *duration) Duration() time.Duration {
	return time.Duration(*du)
}
