package dongauth

import (
	"bytes"
	"io"
	"net/http"
)

type casRequest struct {
	JobNumber   string `json:"jobNumber"`
	CompanyCode string `json:"companyCode"`
	Password    string `json:"password"`
	ClientID    string `json:"clientId"`
}

type uniformResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func (r uniformResponse) succeed() bool {
	return r.Code/100 == 2
}

type DongError struct {
	Code     int
	Message  string
	Body     []byte
	Request  *http.Request
	RawError error
}

func (d *DongError) Error() string {
	buf := new(bytes.Buffer)
	reqURL := d.Request.URL
	buf.WriteString(d.Request.Method)
	buf.WriteString(" ")
	buf.WriteString(reqURL.String())
	buf.WriteString(" ")
	if d.Code != 0 {
		buf.WriteString("认证服务器响应信息：")
		buf.WriteString(d.Message)
	} else {
		buf.WriteString("响应数据为：")
		buf.Write(d.Body)
	}

	return buf.String()
}

func newLimitedWriter(w io.Writer, n int) io.Writer {
	return &limitedWriter{n: n, w: w}
}

type limitedWriter struct {
	n int // remaining bytes allowed
	w io.Writer
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	if l.n <= 0 {
		return 0, nil
	}

	if len(p) > l.n {
		p = p[:l.n]
	}

	n, err := l.w.Write(p)
	l.n -= n

	return n, err
}
