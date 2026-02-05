package service

import (
	"context"
	"log/slog"

	"github.com/vela-ssoc/ssoc-common/datalayer/query"
	"github.com/vela-ssoc/ssoc-manager/integration/dongauth"
)

type Auth struct {
	qry *query.Query
	cli *dongauth.Client
	log *slog.Logger
}

func NewAuth(qry *query.Query, log *slog.Logger) *Auth {
	return &Auth{
		qry: qry,
		log: log,
	}
}

// Qrcode 通过咚咚扫码登录。
func (a *Auth) Qrcode(ctx context.Context, code string) error {
	return nil
}

// CAS 认证登录。
func (a *Auth) CAS(ctx context.Context, uname, passwd string) error {
	attrs := []any{"username", uname}

	// 查询用户是否存在

	// CAS 认证
	if err := a.cli.CAS(ctx, "group", uname, passwd); err != nil {
		attrs = append(attrs, "error", err)
		a.log.Warn("CAS 认证失败", attrs...)
		return err
	}

	return nil
}
