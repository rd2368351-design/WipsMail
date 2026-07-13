package shared

import "math"

type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}

func NewPagination(page, pageSize int) Pagination {
	normalizedPage := page
	normalizedSize := pageSize
	if normalizedPage < 1 {
		normalizedPage = 1
	}
	if normalizedSize < 1 {
		normalizedSize = 10
	}
	if normalizedSize > 100 {
		normalizedSize = 100
	}
	return Pagination{
		Page:     normalizedPage,
		PageSize: normalizedSize,
	}
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func (p Pagination) Limit() int {
	return p.PageSize
}

func (p *Pagination) SetTotal(total int64) {
	p.TotalItems = total
	if total == 0 {
		p.TotalPages = 0
		return
	}
	p.TotalPages = int64(math.Ceil(float64(total) / float64(p.PageSize)))
}

func (p Pagination) HasNext() bool {
	return int64(p.Page) < p.TotalPages
}

func (p Pagination) HasPrev() bool {
	return p.Page > 1
}