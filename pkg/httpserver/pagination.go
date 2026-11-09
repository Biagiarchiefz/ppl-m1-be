package httpserver

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPerPage = 20
	MaxPerPage     = 100
)

type Page struct {
	Page    int
	PerPage int
}

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

// ParsePage membaca `page` dan `per_page`. Nilai di luar batas ditolak (400).
func ParsePage(c *gin.Context) (Page, bool) {
	page, ok1 := positiveInt(c.Query("page"), 1)
	perPage, ok2 := positiveInt(c.Query("per_page"), DefaultPerPage)
	if !ok1 || !ok2 || perPage > MaxPerPage {
		return Page{}, false
	}
	return Page{Page: page, PerPage: perPage}, true
}

func positiveInt(raw string, fallback int) (int, bool) {
	if raw == "" {
		return fallback, true
	}
	n, err := strconv.Atoi(raw)
	return n, err == nil && n >= 1
}

func (p Page) Limit() int32 {
	return int32(p.PerPage)
}

func (p Page) Offset() int32 {
	return int32((p.Page - 1) * p.PerPage)
}

func (p Page) Meta(total int64) Meta {
	perPage := int64(p.PerPage)
	return Meta{
		Page:       p.Page,
		PerPage:    p.PerPage,
		Total:      total,
		TotalPages: (total + perPage - 1) / perPage,
	}
}
