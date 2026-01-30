package muxaccept

import "time"

type AuthRequest struct {
	Secret   string `json:"secret"   validate:"required"` // broker 密钥
	Semver   string `json:"semver"   validate:"required"` // broker 版本号
	Inet     string `json:"inet"     validate:"required"`
	Goos     string `json:"goos"     validate:"required"`
	Goarch   string `json:"goarch"   validate:"required"`
	Hostname string `json:"hostname"`
}

type authResponse struct {
	Code   int         `json:"code"`
	Text   string      `json:"text"`
	Config *BrokConfig `json:"config"`
}

type BrokConfig struct {
	DSN         string        `json:"dsn"`
	MaxOpenConn int           `json:"max_open_conn"`
	MaxIdleConn int           `json:"max_idle_conn"`
	MaxLifeTime time.Duration `json:"max_life_time"`
	MaxIdleTime time.Duration `json:"max_idle_time"`
}

type AuthError struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func (a *AuthError) Error() string {
	return a.Text
}
