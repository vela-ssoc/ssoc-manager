package config

import "crypto/tls"

// Server HTTP 服务相关配置
type Server struct {
	Addr   string   `json:"addr"   yaml:"addr"`                            // 监听地址
	Cert   string   `json:"cert"   yaml:"cert"`                            // 证书
	Pkey   string   `json:"pkey"   yaml:"pkey"`                            // 私钥
	Static string   `json:"static" yaml:"static"`                          // 静态资源路径
	Vhosts []string `json:"vhosts" yaml:"vhosts" validate:"dive,required"` // 虚拟主机
}

func (srv Server) Certs() ([]tls.Certificate, error) {
	if srv.Cert == "" || srv.Pkey == "" {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(srv.Cert, srv.Pkey)
	if err != nil {
		return nil, err
	}
	certs := []tls.Certificate{cert}

	return certs, nil
}
