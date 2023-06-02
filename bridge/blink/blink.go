package blink

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/vela-ssoc/vela-common-mb/problem"
)

type Joiner interface {
	Name() string
	// Auth 开始认证
	Auth(context.Context, Ident) (Issue, http.Header, error)
	// Join 认证成功后接入处理业务逻辑
	Join(net.Conn, Ident, Issue) error
}

type Handler interface {
	http.Handler
	Name() string
}

func New(joiner Joiner) Handler {
	return &blink{
		name:   joiner.Name(),
		joiner: joiner,
	}
}

type blink struct {
	name   string
	joiner Joiner
}

func (bk *blink) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 验证 HTTP 方法
	if method := r.Method; method != http.MethodConnect {
		bk.writeError(w, r, http.StatusBadRequest, "不支持的请求方法：%s", method)
		return
	}

	buf := make([]byte, 100*1024)
	n, _ := io.ReadFull(r.Body, buf)
	var ident Ident
	if err := ident.decrypt(buf[:n]); err != nil {
		bk.writeError(w, r, http.StatusBadRequest, "认证信息错误")
		return
	}

	// 鉴权
	ctx := r.Context()
	issue, header, gex := bk.joiner.Auth(ctx, ident)
	if gex != nil {
		bk.writeError(w, r, http.StatusBadRequest, "认证失败：%s", gex.Error())
		return
	}

	dat, err := issue.encrypt()
	if err != nil {
		bk.writeError(w, r, http.StatusInternalServerError, "内部错误：%s", err.Error())
		return
	}

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		bk.writeError(w, r, http.StatusBadRequest, "客户端连接不可以 Hijack")
		return
	}
	conn, _, jex := hijacker.Hijack()
	if jex != nil {
		bk.writeError(w, r, http.StatusBadRequest, "协议升级失败：%s", jex.Error())
		return
	}

	// -----[ Hijack Successful ]-----

	// 默认规定 http.StatusAccepted 为成功状态码
	code := http.StatusAccepted
	res := &http.Response{
		Status:     http.StatusText(code),
		StatusCode: code,
		Proto:      r.Proto,
		ProtoMajor: r.ProtoMajor,
		ProtoMinor: r.ProtoMinor,
		Header:     header,
		Request:    r,
	}
	if dsz := len(dat); dsz > 0 {
		res.Body = io.NopCloser(bytes.NewReader(dat))
		res.ContentLength = int64(dsz)
	}
	if err = res.Write(conn); err != nil {
		_ = conn.Close()
		return
	}

	if err = bk.joiner.Join(conn, ident, issue); err != nil {
		_ = conn.Close()
	}
}

func (bk *blink) Name() string {
	return bk.name
}

// writeError 写入错误
func (bk *blink) writeError(w http.ResponseWriter, r *http.Request, code int, msg string, args ...string) {
	if len(args) != 0 {
		msg = fmt.Sprintf(msg, args)
	}
	pd := &problem.Detail{
		Type:     bk.name,
		Title:    "认证错误",
		Status:   code,
		Detail:   msg,
		Instance: r.RequestURI,
	}
	_ = pd.JSON(w)
}
