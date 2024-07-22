package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	Client *http.Client
}

func (c Client) JSON(ctx context.Context, rawURL string, header http.Header, result any) error {
	resp, err := c.sendJSON(ctx, http.MethodGet, rawURL, header, nil)
	if err != nil {
		return err
	}
	err = c.unmarshalJSON(resp.Body, result)

	return err
}

func (c Client) PostJSON(ctx context.Context, rawURL string, header http.Header, body, result any) error {
	resp, err := c.sendJSON(ctx, http.MethodPost, rawURL, header, body)
	if err != nil {
		return err
	}
	err = c.unmarshalJSON(resp.Body, result)

	return err
}

func (c Client) PostForm(ctx context.Context, rawURL string, header http.Header, body url.Values, result any) error {
	if header == nil {
		header = make(http.Header, 4)
	}
	header.Set("Content-Type", "application/x-www-form-urlencoded")

	encode := body.Encode()
	req, err := c.newRequest(ctx, http.MethodPost, rawURL, header, strings.NewReader(encode))
	if err != nil {
		return err
	}

	resp, err := c.send(req)
	if err != nil {
		return err
	}

	err = c.unmarshalJSON(resp.Body, result)

	return err
}

func (c Client) Do(req *http.Request) (*http.Response, error) {
	return c.send(req)
}

func (c Client) sendJSON(ctx context.Context, method, rawURL string, header http.Header, body any) (*http.Response, error) {
	if header == nil {
		header = make(http.Header, 4)
	}
	header.Set("Accept", "application/json")

	var r io.Reader
	if method != http.MethodGet && method != http.MethodHead {
		rd, err := c.marshalJSON(body)
		if err != nil {
			return nil, err
		}
		r = rd
		header.Set("Content-Type", "application/json; charset=utf-8")
	}

	req, err := c.newRequest(ctx, method, rawURL, header, r)
	if err != nil {
		return nil, err
	}

	return c.send(req)
}

func (Client) newRequest(ctx context.Context, method, rawURL string, header http.Header, body io.Reader) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	if len(header) != 0 {
		req.Header = header
	}

	return req, nil
}

func (c Client) send(req *http.Request) (*http.Response, error) {
	h := req.Header
	if host := h.Get("Host"); host != "" {
		req.Host = host
	}
	if h.Get("User-Agent") == "" {
		chrome126 := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
		h.Set("User-Agent", chrome126)
	}
	if h.Get("Accept-Language") == "" {
		h.Set("Accept-Language", "zh-CN,zh;q=0.9")
	}

	resp, err := c.getClient().Do(req)
	if err != nil {
		return nil, err
	}

	code := resp.StatusCode
	rem := code / 100
	if rem == 2 || rem == 3 {
		return resp, nil
	}

	e := &Error{
		Code:    code,
		Header:  resp.Header,
		Request: req,
	}
	buf := make([]byte, 1024)
	n, _ := io.ReadFull(resp.Body, buf)
	_ = resp.Body.Close()
	e.Body = buf[:n]

	return nil, e
}

func (c Client) marshalJSON(v any) (io.Reader, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(v)
	return buf, err
}

func (c Client) getClient() *http.Client {
	if cli := c.Client; cli != nil {
		return cli
	}
	return http.DefaultClient
}

func (c Client) unmarshalJSON(rc io.ReadCloser, v any) error {
	//goland:noinspection GoUnhandledErrorResult
	defer rc.Close()
	if v == nil || rc == http.NoBody {
		return nil
	}

	return json.NewDecoder(rc).Decode(v)
}
