package config

import "time"

type Section struct {
	Dong bool          `json:"dong" yaml:"dong"` // 是否发送咚咚验证码
	CDN  string        `json:"cdn"  yaml:"cdn"`  // 文件下载缓存目录
	Sess time.Duration `json:"sess" yaml:"sess"` // session 间隔
}
