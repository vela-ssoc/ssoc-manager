package param

import (
	"github.com/vela-ssoc/ssoc-common-mb/dynsql"
	"gorm.io/gen"
	"gorm.io/gorm"
)

type Pager interface {
	Size() int
	Keyword() string
	Scope(count int64) func(gen.Dao) gen.Dao
	DBScope(count int64) func(*gorm.DB) *gorm.DB
	Result(count int64, records any) *PageResult
}

type Page struct {
	Current int    `query:"current" json:"current"`
	Size    int    `query:"size"    json:"size"`
	Keyword string `query:"keyword" json:"keyword"`
}

func (p Page) Pager() Pager {
	current := p.Current
	size := p.Size
	keyword := p.Keyword
	if current <= 0 {
		current = 1
	}
	if size <= 0 {
		size = 10
	} else if size > 1000 {
		size = 1000
	}
	if keyword != "" {
		keyword = "%" + keyword + "%"
	}
	return &pageScope{
		current: current,
		size:    size,
		keyword: keyword,
	}
}

type pageScope struct {
	current int
	size    int
	keyword string
}

func (ps *pageScope) Size() int {
	return ps.size
}

func (ps *pageScope) Keyword() string {
	return ps.keyword
}

func (ps *pageScope) Scope(count int64) func(gen.Dao) gen.Dao {
	if count > 0 {
		size := int64(ps.size)
		page := int((count + size - 1) / size)
		if page < ps.current {
			ps.current = page
		}
	}
	return func(dao gen.Dao) gen.Dao {
		return dao.Offset((ps.current - 1) * ps.size).Limit(ps.size)
	}
}

func (ps *pageScope) DBScope(count int64) func(*gorm.DB) *gorm.DB {
	if count > 0 {
		size := int64(ps.size)
		page := int((count + size - 1) / size)
		if page < ps.current {
			ps.current = page
		}
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset((ps.current - 1) * ps.size).Limit(ps.size)
	}
}

func (ps *pageScope) Result(count int64, records any) *PageResult {
	ret := &PageResult{
		Total:   count,
		Current: ps.current,
		Size:    ps.size,
		Records: records,
	}
	if count <= 0 || records == nil {
		ret.Records = []struct{}{}
	}
	return ret
}

type PageResult struct {
	Total   int64 `json:"total"`
	Current int   `json:"current"`
	Size    int   `json:"size"`
	Records any   `json:"records"`
}

type Index struct {
	Size    int    `query:"size"`
	Keyword string `query:"keyword"`
}

func (i Index) Indexer() Indexer {
	size, keyword := i.Size, i.Keyword
	if size <= 0 || size > 1000 {
		size = 1000
	}
	if keyword != "" {
		keyword = "%" + keyword + "%"
	}
	return &indexScope{
		size:    size,
		keyword: keyword,
	}
}

type Indexer interface {
	Size() int
	Keyword() string
	Scope(gen.Dao) gen.Dao
	Result(dats any) *IndexResult
}

type indexScope struct {
	size    int
	keyword string
}

func (is *indexScope) Size() int {
	return is.size
}

func (is *indexScope) Keyword() string {
	return is.keyword
}

func (is *indexScope) Scope(dao gen.Dao) gen.Dao {
	return dao.Limit(is.size)
}

func (is *indexScope) Result(dats any) *IndexResult {
	if dats == nil {
		dats = []struct{}{}
	}
	return &IndexResult{
		Records: dats,
	}
}

type IndexResult struct {
	Records any
}

type PageSQL struct {
	Page
	dynsql.Input
}
