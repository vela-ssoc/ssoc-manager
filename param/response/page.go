package response

import "gorm.io/gen"

type Pages[E any] struct {
	Page    int64 `json:"page"`    // 页码
	Size    int64 `json:"size"`    // 每页显示条数
	Total   int64 `json:"total"`   // 总条数
	Records []E   `json:"records"` // 本页数据
}

func NewPages[E any](page, size int64) *Pages[E] {
	return &Pages[E]{
		Page:    page,
		Size:    size,
		Records: []E{},
	}
}

// FP 回退分页（Fallback Pagination），当用户请求一个不存在的页码时，
// 不会看到空白或错误提示，而是会看到有效的内容（最后一页数据）， 从而
// 提高用户体验。
func (p *Pages[E]) FP(cnt int64) func(dao gen.Dao) gen.Dao {
	return func(dao gen.Dao) gen.Dao {
		page, size := p.Page, p.Size
		if page <= 0 {
			page = 1
		}
		if size <= 0 {
			size = 10
		} else if size > 1000 {
			size = 1000
		}
		if cnt < 0 {
			cnt = 0
		}

		if pg := (cnt + size - 1) / size; pg < page {
			page = pg
		}
		p.Page, p.Size, p.Total = page, size, cnt
		offset, limit := int((page-1)*size), int(size)

		return dao.Offset(offset).Limit(limit)
	}
}

func (p *Pages[E]) Empty() *Pages[E] {
	p.Page, p.Total, p.Records = 1, 0, []E{}
	return p
}

func (p *Pages[E]) SetRecords(es []E) *Pages[E] {
	if es == nil {
		es = []E{}
	}
	p.Records = es

	return p
}
