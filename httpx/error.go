package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Error struct {
	Code    int
	Header  http.Header
	Body    []byte
	Request *http.Request
}

func (e *Error) Error() string {
	var method string
	var rawURL *url.URL
	if req := e.Request; req != nil {
		method = req.Method
		rawURL = req.URL
	}
	return fmt.Sprintf("http client failed %s %s: %d %s", method, rawURL, e.Code, e.Body)
}

func (e *Error) JSON(v any) error {
	return json.Unmarshal(e.Body, v)
}
