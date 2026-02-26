package errcode

import "github.com/xgfone/ship/v5"

var (
	ErrNilDocument      = ship.ErrBadRequest.Newf("数据不存在")
	ErrCertificateParse = ship.ErrBadRequest.Newf("证书解析失败")
	ErrDataConflict     = ship.ErrStatusConflict.Newf("资源已存在")
)
