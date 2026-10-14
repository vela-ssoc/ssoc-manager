package request

import (
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Size struct {
	Size int64 `json:"size" query:"size" form:"size" validate:"lte=1000"`
}

func (s Size) SN() int64 {
	if s.Size <= 0 {
		return 10
	} else if s.Size > 1000 {
		return 1000
	}

	return s.Size
}

type Page struct {
	Page int64 `json:"page" query:"page" form:"page" validate:"gte=0"`
}

func (p Page) PN() int64 {
	if p.Page <= 0 {
		return 1
	}

	return p.Page
}

type PageSize struct {
	Page
	Size
}

func (p PageSize) PSN() (int64, int64) {
	return p.PN(), p.SN()
}

type PageSizeKeyword struct {
	PageSize
	Keyword
}

type Keyword struct {
	Q string `json:"q" query:"q" validate:"lte=100"`
}

func (k Keyword) Trim() string {
	return strings.TrimSpace(k.Q)
}

func (k Keyword) Regex(fields ...string) bson.A {
	if len(fields) == 0 {
		return nil
	}
	q := k.Trim()
	if q == "" {
		return nil
	}

	pat := regexp.QuoteMeta(q)
	reg := &bson.Regex{Pattern: pat, Options: "i"}
	arr := make(bson.A, 0, len(fields))
	for _, f := range fields {
		arr = append(arr, bson.D{{Key: f, Value: reg}})
	}

	return arr
}

func (k Keyword) RegexTo(dst bson.M, fields ...string) bson.M {
	if dst == nil {
		dst = make(bson.M, 4)
	}
	if arr := k.Regex(fields...); len(arr) != 0 {
		dst["$or"] = arr
	}

	return dst
}
