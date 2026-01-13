package request

import "gorm.io/gen"

type Pages struct {
	Page int `query:"page" json:"page" form:"page"`
	Size int `query:"size" json:"size" form:"size" validate:"lte=1000"`
}

func (p Pages) PageSize() (page, size int) {
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

func (p Pages) Scope(dao gen.Dao) gen.Dao {
	page, size := p.PageSize()
	return dao.Offset((page - 1) * size).Limit(size)
}
