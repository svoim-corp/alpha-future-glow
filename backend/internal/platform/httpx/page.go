package httpx

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	defaultPage = 1
	defaultSize = 20
	maxSize     = 100
)

// Page — обёртка списка с пагинацией (API_SPEC.md, раздел 3).
type Page[T any] struct {
	Items []T `json:"items"`
	Page  int `json:"page"`
	Size  int `json:"size"`
	Total int `json:"total"`
}

// NewPage гарантирует, что items не null.
func NewPage[T any](items []T, page, size, total int) Page[T] {
	if items == nil {
		items = []T{}
	}
	return Page[T]{Items: items, Page: page, Size: size, Total: total}
}

// ParsePage читает ?page и ?size: по умолчанию 1/20, size не больше 100.
func ParsePage(c *gin.Context) (page, size int) {
	page, size = defaultPage, defaultSize
	if n, err := strconv.Atoi(c.Query("page")); err == nil && n >= 1 {
		page = n
	}
	if n, err := strconv.Atoi(c.Query("size")); err == nil && n >= 1 {
		size = min(n, maxSize)
	}
	return page, size
}
