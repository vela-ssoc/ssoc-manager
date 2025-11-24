package serverd

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/vela-ssoc/ssoc-common/linkhub"
)

func defaultValid(v any) error {
	req, ok := v.(*authRequest)
	if !ok {
		return errors.New("认证报文结构体无效")
	}
	if req.Secret == "" {
		return errors.New("连接密钥必须填写")
	}
	if req.Semver == "" {
		return errors.New("版本号必须填写")
	}

	return nil
}

type Limiter interface {
	Allowed() bool
}

type unlimited struct{}

func (*unlimited) Allowed() bool { return true }

type option struct {
	logger   *slog.Logger
	valid    func(any) error
	server   *http.Server
	limit    Limiter
	huber    linkhub.Huber
	timeout  time.Duration
	notifier BrokerNotifier
}

func NewOption() OptionBuilder {
	return OptionBuilder{}
}

type OptionBuilder struct {
	opts []func(option) option
}

func (ob OptionBuilder) List() []func(option) option {
	return ob.opts
}

func (ob OptionBuilder) Logger(v *slog.Logger) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.logger = v
		return o
	})
	return ob
}

func (ob OptionBuilder) Valid(v func(any) error) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.valid = v
		return o
	})
	return ob
}

func (ob OptionBuilder) Server(v *http.Server) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.server = v
		return o
	})
	return ob
}

func (ob OptionBuilder) Handler(v http.Handler) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		if o.server == nil {
			o.server = new(http.Server)
		}
		o.server.Handler = v

		return o
	})
	return ob
}

func (ob OptionBuilder) Limit(v Limiter) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.limit = v
		return o
	})
	return ob
}

func (ob OptionBuilder) Timeout(v time.Duration) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.timeout = v
		return o
	})
	return ob
}

func (ob OptionBuilder) Huber(v linkhub.Huber) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.huber = v
		return o
	})
	return ob
}

func (ob OptionBuilder) BrokerNotifier(v BrokerNotifier) OptionBuilder {
	ob.opts = append(ob.opts, func(o option) option {
		o.notifier = v
		return o
	})
	return ob
}

func fallbackOption() OptionBuilder {
	return OptionBuilder{
		opts: []func(option) option{
			func(o option) option {
				if o.valid == nil {
					o.valid = defaultValid
				}
				if o.server == nil {
					o.server = new(http.Server)
				}
				if o.server.Handler == nil {
					o.server.Handler = http.NotFoundHandler()
				}
				if o.limit == nil {
					o.limit = new(unlimited)
				}
				if o.huber == nil {
					o.huber = linkhub.NewSafeMap()
				}
				if o.timeout <= 0 {
					o.timeout = 30 * time.Second
				}
				if o.notifier == nil {
					o.notifier = new(brokerNotifier)
				}

				return o
			},
		},
	}
}
