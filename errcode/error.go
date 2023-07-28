package errcode

import "github.com/xgfone/ship/v5"

var (
	ErrUnauthorized = ship.ErrUnauthorized.Newf("认证无效")
	ErrForbidden    = ship.ErrForbidden.Newf("禁止操作")

	ErrUnsupportedWebSocket = ship.ErrBadRequest.Newf("该接口接口不支持 websocket 请求")
	ErrRequiredWebSocket    = ship.ErrBadRequest.Newf("该接口必须是 websocket 协议的请求")
	ErrNodeNotExist         = ship.ErrBadRequest.Newf("节点不存在或已经离线")
	ErrTagNotExist          = ship.ErrBadRequest.Newf("标签不存在")
	ErrSubstanceNotExist    = ship.ErrBadRequest.Newf("配置不存在")
	ErrSubstanceEffected    = ship.ErrBadRequest.Newf("配置已经发布")
	ErrRequiredNode         = ship.ErrBadRequest.Newf("节点信息必须填写")
	ErrRequiredAddr         = ship.ErrBadRequest.Newf("地址必须填写")
	ErrPictureCode          = ship.ErrBadRequest.Newf("图片验证码错误")
	ErrDongCode             = ship.ErrBadRequest.Newf("咚咚验证码错误")
	ErrWithoutDongCode      = ship.ErrBadRequest.Newf("无需发送咚咚验证码")
	ErrTooManyLoginFailed   = ship.ErrBadRequest.Newf("登录错误次数较多")
	ErrPassword             = ship.ErrBadRequest.Newf("密码错误")
	ErrTaskBusy             = ship.ErrBadRequest.Newf("当前任务繁忙")
	ErrVersion              = ship.ErrBadRequest.Newf("请刷新后操作")
	ErrRequiredFilter       = ship.ErrBadRequest.Newf("该操作至少包含一个过滤条件")
	ErrRequiredGroup        = ship.ErrBadRequest.Newf("group 条件必须填写")
	ErrDeleteFailed         = ship.ErrBadRequest.Newf("删除失败")
	ErrCertMatchKey         = ship.ErrBadRequest.Newf("证书与私钥不匹配")
	ErrCertUsedByBroker     = ship.ErrBadRequest.Newf("证书已被代理节点使用")
	ErrCertificate          = ship.ErrBadRequest.Newf("证书错误")
	ErrDeleteSelf           = ship.ErrBadRequest.Newf("禁止删除自己")
	ErrOperateFailed        = ship.ErrBadRequest.Newf("操作失败")
	ErrNodeStatus           = ship.ErrBadRequest.Newf("节点状态不允许操作")
	ErrExceedAuthority      = ship.ErrBadRequest.Newf("越权访问")
	ErrDeprecated           = ship.ErrBadRequest.Newf("版本已被标记为过期")
	ErrInetAddress          = ship.ErrBadRequest.Newf("inet 地址无效")
	ErrAlreadyExist         = ship.ErrBadRequest.Newf("数据已存在")
	ErrInvalidData          = ship.ErrBadRequest.Newf("数据验证无效")
)

type Errorf interface {
	Fmt(...any) error
}

type formatError string

func (f formatError) Fmt(a ...any) error {
	return ship.ErrBadRequest.Newf(string(f), a...)
}

const (
	FmtErrNameExist = formatError("名字 %s 已经存在")
	FmtErrInetExist = formatError("inet %s 已经存在")
	FmtErrTaskBusy  = formatError("任务繁忙：%d")
)
