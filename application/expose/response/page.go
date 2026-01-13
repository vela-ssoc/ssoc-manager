package response

import "gorm.io/gen"

type Pages[E any] struct {
	Page    int   `json:"page"`    // 页码
	Size    int   `json:"size"`    // 每页显示条数
	Total   int64 `json:"total"`   // 总条数
	Records []*E  `json:"records"` // 数据
}

func (p *Pages[E]) PageSize() (page, size int) {
	page, size = p.Page, p.Size
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	} else if size > 1000 {
		size = 1000
	}

	return
}

func (p *Pages[E]) SetRecords(records []*E) *Pages[E] {
	if records == nil {
		records = []*E{} // 返回 [] 而不是 null
	}
	p.Records = records

	return p
}

func (p *Pages[E]) Scope(total int64) func(dao gen.Dao) gen.Dao {
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

		if last := (int(total) + size - 1) / size; last < page {
			page = last
		}

		p.Page, p.Size, p.Total = page, size, total
		offset, limit := int((page-1)*size), int(size)

		return dao.Offset(offset).Limit(limit)
	}
}

func NewPages[E any](page, size int) *Pages[E] {
	return &Pages[E]{
		Page:    page,
		Size:    size,
		Records: []*E{}, // 返回 [] 而不是 null
	}
}
