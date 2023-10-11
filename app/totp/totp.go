package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TOTP https://github.com/google/google-authenticator/wiki/Key-Uri-Format
//
//	FIXME 经过测试以下几款 TOTP 程序，均不能完全兼容上述规则，其中 T盾动态密码 兼容性最差。\
//		Google Authenticator、Microsoft Authenticator、FreeOTP、数盾OTP（微信小程序）、T盾动态密码（微信小程序），
//		为了提高各个软件对 TOTP 的兼容性，最好将算法、动态码位数、动态码刷新时间保持默认值。
type TOTP struct {
	Issuer    string `json:"issuer"    xml:"issuer"    yaml:"issuer"`    // 可选：签发者
	Account   string `json:"account"   xml:"account"   yaml:"account"`   // 可选：账户
	Secret    string `json:"secret"    xml:"secret"    yaml:"secret"`    // 必选：Base32 编码后的密钥
	Algorithm string `json:"algorithm" xml:"algorithm" yaml:"algorithm"` // 可选：合法值 SHA1/SHA256/SHA512 默认 SHA1
	Digits    int    `json:"digits"    xml:"digits"    yaml:"digits"`    // 可选：合法值 6/8 默认 6
	Period    int    `json:"period"    xml:"period"    yaml:"period"`    // 可选：合法值 15/30/60 默认 30
	Image     string `json:"image"     xml:"image"     yaml:"image"`     // 非标：部分软件支持自定义 http 图标，如 FreeOTP
}

// String 生成 OTP URL.
//
//	FIXME: OTP 中 label 字段应该使用 url-encode 编码，但是在 Go 语言中的 url.QueryEscape 和 url.PathEscape 都不是
//		完全意义上的 url-encode，例如：英文空格在 url-encode 后是 %20 但是 Go url.QueryEscape 却转义成了英文加号；
//		英文冒号在 url-encode 中应该是 %3A 但是 Go 的 url.PathEscape 却不做转义。
func (t TOTP) String() string {
	return t.URL().String()
}

func (t TOTP) URL() *url.URL {
	quires := make(url.Values, 8)
	quires["secret"] = []string{t.Secret}
	if issuer := t.Issuer; issuer != "" {
		quires.Set("issuer", issuer)
	}
	if alg := strings.ToUpper(t.Algorithm); alg == "SHA256" || alg == "SHA512" {
		quires.Set("algorithm", alg)
	}
	if digit := t.Digits; digit == 8 {
		digits := strconv.Itoa(digit)
		quires.Set("digits", digits)
	}
	if second := t.Period; second == 15 || second == 60 {
		period := strconv.Itoa(second)
		quires.Set("period", period)
	}
	// image 一般是 HTTP 图片链接，FreeOTP 会解析该参数作为图标，
	// 但是该参数会导致 [数盾OTP] 无法对 issuer 和 account 解析。
	if img := t.Image; img != "" {
		quires.Set("image", img)
	}

	return &url.URL{
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     t.Issuer + ":" + t.Account,
		RawQuery: quires.Encode(),
	}
}

// Generate 随机生成默认参数的 TOTP
func Generate(issuer, account string) *TOTP {
	rb := make([]byte, 30)
	_, _ = rand.Read(rb)
	secret := base32.StdEncoding.
		WithPadding(base32.NoPadding).
		EncodeToString(rb)

	return &TOTP{
		Issuer:  issuer,
		Account: account,
		Secret:  secret,
	}
}

func Validate(secret, code string, strict bool) bool {
	if secret == "" || code == "" {
		return false
	}

	const period, ago = 30, 10
	now := time.Now().Unix()
	cur := now / period
	if otp, err := calculateCode(secret, cur); err == nil && otp == code {
		return true
	}
	if strict {
		return false
	}
	last := (now - ago) / period
	if last == cur {
		return false
	}

	otp, err := calculateCode(secret, last)

	return err == nil && otp == code
}

func calculateCode(secret string, rem int64) (string, error) {
	b32, err := base32.StdEncoding.
		WithPadding(base32.NoPadding).
		DecodeString(secret)
	if err != nil {
		return "", err
	}

	datetime := make([]byte, 8)
	binary.BigEndian.PutUint64(datetime, uint64(rem))

	h := hmac.New(sha1.New, b32)
	h.Write(datetime)
	sum := h.Sum(nil)

	offset := sum[len(sum)-1] & 0x0f
	truncated := binary.BigEndian.Uint32(sum[offset:]) & 0x7fffffff
	code := truncated % 1_000_000

	return fmt.Sprintf("%06d", code), nil
}
