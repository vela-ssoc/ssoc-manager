package dongauth

import (
	"bytes"
	"context"
	"crypto/aes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type CASConfig struct {
	// URL 请求地址。
	// 例如：https://example.com/dongdong-auth/open/login
	URL      string
	ClientID string
	Secret   string
}

type Configurer interface {
	CASConfig(ctx context.Context) (*CASConfig, error)
}

type Client struct {
	cfg Configurer
	cli *http.Client
	log *slog.Logger
}

func NewClient(cfg Configurer, cli *http.Client, log *slog.Logger) *Client {
	return &Client{
		cfg: cfg,
		cli: cli,
		log: log,
	}
}

func (c *Client) CAS(ctx context.Context, company, jobNumber, password string) error {
	cfg, err := c.cfg.CASConfig(ctx)
	if err != nil {
		return err
	}
	passwd, err := c.encryptECBPKCS7(password, cfg.Secret)
	if err != nil {
		return err
	}

	req := &casRequest{
		JobNumber:   jobNumber,
		CompanyCode: company,
		Password:    passwd,
		ClientID:    cfg.ClientID,
	}
	err = c.sendJSON(ctx, http.MethodPost, cfg.URL, req, nil)

	return err
}

func (c *Client) sendJSON(ctx context.Context, method, strURL string, body, result any) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, strURL, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}
	rc := resp.Body
	//goland:noinspection GoUnhandledErrorResult
	defer rc.Close()

	const recordSize = 4096
	code := resp.StatusCode
	if code/100 != 2 { // 状态码错误。
		chk := make([]byte, recordSize)
		n, _ := io.ReadFull(rc, chk)

		return &DongError{
			Body:    chk[:n],
			Request: req,
		}
	}

	record := new(bytes.Buffer)
	lw := newLimitedWriter(record, recordSize)
	rd := io.TeeReader(rc, lw)

	res := &uniformResponse{Data: result}
	err = json.NewDecoder(rd).Decode(res)
	if err == nil && res.succeed() {
		return nil
	}

	return &DongError{
		Code:     res.Code,
		Message:  res.Message,
		Body:     record.Bytes(),
		Request:  req,
		RawError: err,
	}
}

// encryptECBPKCS7 AES/ECB/PKCS7Padding
func (*Client) encryptECBPKCS7(plaintext, secret string) (string, error) {
	key, err := hex.DecodeString(secret)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// PKCS7Padding
	padded := []byte(plaintext)
	blockSize := block.BlockSize()
	padding := blockSize - len(padded)%blockSize
	pads := bytes.Repeat([]byte{byte(padding)}, padding)
	padded = append(padded, pads...)

	ciphertext := make([]byte, len(padded))
	for bs, be := 0, blockSize; bs < len(padded); bs, be = bs+blockSize, be+blockSize {
		block.Encrypt(ciphertext[bs:be], padded[bs:be])
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
