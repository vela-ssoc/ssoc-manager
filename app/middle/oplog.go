package middle

import (
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-manager/app/route"
	"github.com/vela-ssoc/ssoc-manager/app/session"
	"github.com/xgfone/ship/v5"
)

type OplogSaver interface{}

// Oplog 操作日志记录中间件，内部包含 recovery，无需额外添加 recovery 中间件
func Oplog(recd route.Recorder) ship.Middleware {
	newFn := func() any {
		// 最多记录前 1K 的报文数据
		const maxsize = 1024
		return &limitCopy{
			maxsize: maxsize,
			data:    make([]byte, maxsize),
		}
	}
	m := &oplogMid{
		recd: recd,
		pool: sync.Pool{New: newFn},
	}

	return m.middleware
}

type oplogMid struct {
	recd route.Recorder
	pool sync.Pool
}

func (m *oplogMid) middleware(handler ship.Handler) ship.Handler {
	return func(c *ship.Context) error {
		w, r := c.Response(), c.Request()
		ctx := r.Context()
		body := m.getLimitCopy(r.Body)
		r.Body = body

		reqURL, length := c.Request().URL, c.ContentLength()
		addr, method := c.RemoteAddr(), c.Method()
		path, query, clientIP := reqURL.Path, reqURL.RawQuery, c.ClientIP()

		begin := time.Now()
		err := handler(c)
		elapsed := time.Since(begin)

		ext, ok := c.Route.Data.(route.Describer)
		if ok && ext.Ignore() {
			return err
		}

		data := body.bytes()
		oplog := &model.Oplog{
			ClientAddr: clientIP,
			DirectAddr: addr,
			Method:     method,
			Path:       path,
			Query:      query,
			Length:     length,
			Content:    data,
			RequestAt:  begin,
			Elapsed:    elapsed,
		}
		if ext != nil {
			oplog.Name = ext.Name(c)
			oplog.Content = ext.Desensitization(data)
		}

		if err != nil {
			oplog.Failed, oplog.Cause = true, err.Error()
		} else if code := w.Status; code >= http.StatusBadRequest {
			oplog.Failed, oplog.Cause = true, strconv.Itoa(code)
		}
		if info := session.Cast(c.Any); info != nil {
			oplog.UserID = info.ID
			oplog.Username = info.Username
			oplog.Nickname = info.Nickname
		}
		_ = m.recd.Save(ctx, oplog)

		return err
	}
}

func (m *oplogMid) getLimitCopy(body io.ReadCloser) *limitCopy {
	lc := m.pool.Get().(*limitCopy)
	lc.body = body
	lc.pos = 0
	return lc
}

// limitCopy 带读取限制的 http request body
// 日志中间件需要记录原生请求 Body，但是 Body 有长有短，对与过长
// 的请求我们只记录前 max 位数据，后续的数据不再记录。
type limitCopy struct {
	body    io.ReadCloser // 原来的 HTTP Body
	maxsize int           // 最大记录的 Body 长度
	pos     int           // 已记录的数据偏移量
	data    []byte        // 数据
}

func (lc *limitCopy) Read(p []byte) (int, error) {
	n, err := lc.body.Read(p)
	if n != 0 && lc.maxsize > lc.pos {
		num := copy(lc.data[lc.pos:], p[:n])
		lc.pos += num
	}
	return n, err
}

func (lc *limitCopy) Close() error {
	return lc.body.Close()
}

func (lc *limitCopy) bytes() []byte {
	return lc.data[:lc.pos]
}
