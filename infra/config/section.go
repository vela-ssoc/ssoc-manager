package config

import (
	"context"
	"time"

	"github.com/vela-ssoc/vela-common-mb/cmdb2"
)

type Section struct {
	Dong  bool          `json:"dong"  yaml:"dong"` // 是否发送咚咚验证码
	CDN   string        `json:"cdn"   yaml:"cdn"`  // 文件下载缓存目录
	Sess  time.Duration `json:"sess"  yaml:"sess"` // session 间隔
	Cmdb2 Cmdb2         `json:"cmdb2" yaml:"cmdb2"`
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
