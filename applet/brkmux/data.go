package brkmux

import (
	"encoding/json"
	"net"
	"time"

	"github.com/vela-ssoc/vela-common-mb/dbms"
	"github.com/vela-ssoc/vela-common-mba/ciphertext"
	"github.com/vela-ssoc/vela-manager/profile"
)

// Ident broker 节点的认证信息
type Ident struct {
	ID         int64     `json:"id"`         // ID
	Secret     string    `json:"secret"`     // 密钥
	Inet       net.IP    `json:"inet"`       // IPv4 地址
	MAC        string    `json:"mac"`        // MAC 地址
	Semver     string    `json:"semver"`     // 版本
	Goos       string    `json:"goos"`       // runtime.GOOS
	Arch       string    `json:"arch"`       // runtime.GOARCH
	CPU        int       `json:"cpu"`        // runtime.NumCPU
	PID        int       `json:"pid"`        // os.Getpid
	Workdir    string    `json:"workdir"`    // os.Getwd
	Executable string    `json:"executable"` // os.Executable
	Username   string    `json:"username"`   // user.Current
	Hostname   string    `json:"hostname"`   // os.Hostname
	TimeAt     time.Time `json:"time_at"`    // 发起时间
}

// String fmt.Stringer
func (ide Ident) String() string {
	dat, _ := json.MarshalIndent(ide, "", "    ")
	return string(dat)
}

// decrypt 将数据解密至 Ident
func (ide *Ident) decrypt(enc []byte) error {
	return ciphertext.DecryptJSON(enc, ide)
}

// Issue broker 节点认证成功后返回的信息
type Issue struct {
	Name     string         `json:"name"`     // broker 名字
	Passwd   []byte         `json:"passwd"`   // 通信加解密密钥
	Listen   Listen         `json:"listen"`   // 服务监听配置
	Logger   profile.Logger `json:"logger"`   // 日志配置
	Database dbms.Config    `json:"database"` // 数据库配置
	Section  Section        `json:"section"`
}

// String fmt.Stringer
func (iss Issue) String() string {
	dat, _ := json.MarshalIndent(iss, "", "    ")
	return string(dat)
}

// encrypt 将 Issue 加密
func (iss Issue) encrypt() ([]byte, error) {
	return ciphertext.EncryptJSON(iss)
}

// Listen 监听信息
type Listen struct {
	Addr string `json:"addr"` // 监听地址 :8080 192.168.1.2:8080
	Cert string `json:"cert"` // 证书
	Pkey string `json:"pkey"` // 私钥
}

type Section struct {
	CDN string `json:"cdn"` // 文件下载缓存目录
}
