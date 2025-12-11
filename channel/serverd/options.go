package serverd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vela-ssoc/ssoc-common/linkhub"
	"github.com/vela-ssoc/vela-common-mba/smux"
)

type BootConfigLoader interface {
	LoadBootConfig(ctx context.Context) (*BootConfig, error)
}

type BrokerEventer interface {
	OnConnected(linkhub.Peer)
	OnDisconnected(linkhub.Peer)
}

type Options struct {
	Handler     http.Handler             // 必填
	Huber       linkhub.Huber            // 必填
	BootConfig  BootConfigurer           // 必填
	Logger      *slog.Logger             // 建议填
	Valid       func(any) error          // 建议填
	Allow       func(*smux.Session) bool // 限流器
	Timeout     time.Duration
	BrokerEvent BrokerEventer
}

func (o Options) precheck() error {
	if o.Handler == nil {
		return errors.New("请配置处理程序 (Options.Handler)")
	}
	if o.Huber == nil {
		return errors.New("请配置连接池 (Options.Huber)")
	}
	if o.BootConfig == nil {
		return errors.New("请配置初始化配置加载方式 (Options.BootConfig)")
	}

	return nil
}

func (o Options) logger() *slog.Logger {
	if log := o.Logger; log != nil {
		return log
	}

	return slog.Default()
}

func (o Options) timeout() time.Duration {
	if du := o.Timeout; du > 0 {
		return du
	}

	return 30 * time.Second
}

func (o Options) brokerEvent() BrokerEventer {
	if be := o.BrokerEvent; be != nil {
		return be
	}

	return noopBrokerEvent{}
}

func (o Options) valid(v any) error {
	if valid := o.Valid; valid != nil {
		return valid(v)
	}

	req, ok := v.(*authRequest)
	if !ok {
		return fmt.Errorf("不符合预期的请求类型 (%T)", v)
	}
	if req.Secret == "" {
		return errors.New("连接密钥必须填写 (secret)")
	}
	if req.Semver == "" {
		return errors.New("版本号必须填写 (semver)")
	}

	return nil
}

func (o Options) allow() bool {
}

type noopBrokerEvent struct{}

func (noopBrokerEvent) OnConnected(linkhub.Peer)    {}
func (noopBrokerEvent) OnDisconnected(linkhub.Peer) {}
