package modview

import (
	"net/http"
	"time"
)

// LoginDong 登录发送咚咚验证码时，提供的模型参数用于模板渲染。
type LoginDong struct {
	Header   http.Header `json:"header"`    // 请求 Header 信息
	Username string      `json:"username"`  // 登录用户名
	Nickname string      `json:"nickname"`  // 用户昵称
	RemoteIP string      `json:"remote_ip"` // 登录远程地址
	ClientIP string      `json:"client_ip"` // 直连地址
	LoginAt  time.Time   `json:"login_at"`  // 登录时间
	Code     string      `json:"code"`      // 咚咚验证码
	Minute   int         `json:"minute"`    // 验证码有效分钟数，一般是 1-10
}
