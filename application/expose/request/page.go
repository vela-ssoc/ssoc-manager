package request

import (
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Pages struct {
	Page int64 `json:"page" query:"page" form:"page" validate:"gte=0"`
	Size int64 `json:"size" query:"size" form:"size" validate:"gte=0,lte=1000"`
}

type Keywords struct {
	Keyword string `json:"keyword" query:"keyword" form:"keyword"`
}

func (k Keywords) Regexp(fields ...string) bson.A {
	return k.Regexps(fields)
}

func (k Keywords) Regexps(fields []string) bson.A {
	if len(fields) == 0 {
		return nil
	}

	kw := strings.TrimSpace(k.Keyword)
	if kw == "" {
		return nil
	}

	reg := bson.Regex{Pattern: kw, Options: "i"}
	arr := make(bson.A, 0, len(fields))
	for _, f := range fields {
		arr = append(arr, bson.D{{Key: f, Value: reg}})
	}

	return arr
}

type PageKeywords struct {
	Pages
	Keywords
}
