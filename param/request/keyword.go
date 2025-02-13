package request

import (
	"strings"

	"gorm.io/gen"
	"gorm.io/gen/field"
)

type Keywords struct {
	Keyword string `json:"keyword" query:"keyword" form:"keyword" validate:"max=255"`
}

func (k Keywords) Likes(fields ...field.String) func(dao gen.Dao) gen.Dao {
	return func(dao gen.Dao) gen.Dao {
		kw := strings.TrimSpace(k.Keyword)
		if kw == "" {
			return dao
		}

		kw = "%" + kw + "%"
		likes := make([]field.Expr, 0, len(fields))
		for _, f := range fields {
			likes = append(likes, f.Regexp(kw))
		}

		return dao.Where(field.Or(likes...))
	}
}

func (k Keywords) Regexps(fields ...field.String) func(dao gen.Dao) gen.Dao {
	return func(dao gen.Dao) gen.Dao {
		kw := strings.TrimSpace(k.Keyword)
		if kw == "" {
			return dao
		}

		likes := make([]field.Expr, 0, len(fields))
		for _, f := range fields {
			likes = append(likes, f.Regexp(kw))
		}
		if len(likes) == 0 {
			return dao
		}

		return dao.Where(field.Or(likes...))
	}
}
