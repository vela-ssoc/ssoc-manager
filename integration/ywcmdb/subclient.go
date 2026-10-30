package ywcmdb

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type subClient[S ~[]E, E any] struct {
	cfg     ConfigLoader
	cli     *http.Client
	log     *slog.Logger
	ciTypes []string
	size    int
	ctx     context.Context
}

func newSubClient[S ~[]E, E any](ctx context.Context, ciTypes []string, size int, cc *cmdbClient) *subClient[S, E] {
	return &subClient[S, E]{
		cfg:     cc.cfg,
		cli:     cc.cli,
		log:     cc.log,
		ciTypes: ciTypes,
		size:    size,
		ctx:     ctx,
	}
}

func (sc *subClient[S, E]) All() iter.Seq2[*ResponseBody[S, E], error] {
	return func(yield func(*ResponseBody[S, E], error) bool) {
		ctx, size := sc.ctx, sc.size
		if size <= 0 {
			size = 100
		} else if size > 1000 {
			size = 1000
		}

		cfg, err := sc.cfg.LoadCMDBConfig(ctx)
		if err != nil {
			yield(nil, err)
			sc.log.ErrorContext(ctx, "加载 CMDB 配置出错", "err", err)
			return
		}

		rawURL, err := url.Parse(cfg.Endpoint)
		if err != nil {
			yield(nil, err)
			sc.log.ErrorContext(ctx, "CMDB URL 无效", "url", cfg.Endpoint, "err", err)
			return
		}
		if rawURL.Path == "" || rawURL.Path == "/" {
			sc.log.DebugContext(ctx, "自动补全 CMDB 请求路径")
			rawURL = rawURL.JoinPath("/cmdb2/api/v0.1/ci/s")
		}
		query := rawURL.Query()
		query.Set("_key", cfg.Key)
		query.Set("_secret", cfg.Secret)
		query.Set("q", "_type:("+strings.Join(sc.ciTypes, ";")+")")
		query.Set("count", strconv.Itoa(size))

		var page int
		var cursor int
		for {
			page++
			query.Set("page", strconv.Itoa(page))
			rawURL.RawQuery = query.Encode()
			strURL := rawURL.String()

			resp, err1 := sc.query(ctx, strURL)
			if err1 != nil {
				yield(nil, err1)
				return
			}

			found := resp.NumFound
			num := len(resp.Result)
			cursor += num

			if !yield(resp, nil) {
				return
			}

			if cursor >= found { // 防止服务端：超页返回最后一页
				return
			}
		}
	}
}

func (sc *subClient[S, E]) query(ctx context.Context, strURL string) (*ResponseBody[S, E], error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strURL, nil)
	if err != nil {
		return nil, err
	}

	cli := sc.getClient()
	res, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	body := res.Body
	defer body.Close()

	code := res.StatusCode
	if code/100 != 2 { // != 2xx
		buf := make([]byte, 1024)
		n, _ := io.ReadFull(body, buf)
		he := &HTTPError{
			Code:    code,
			Body:    buf[:n],
			Request: req,
		}

		return nil, he
	}

	result := new(ResponseBody[S, E])
	if err = json.NewDecoder(body).Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (sc *subClient[S, E]) getClient() *http.Client {
	if sc.cli != nil {
		return sc.cli
	}

	return http.DefaultClient
}
