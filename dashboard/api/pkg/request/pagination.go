package request

import (
	"math"
	"strconv"

	"github.com/labstack/echo/v5"
)

// Pagination represents the pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// PaginatorResponse is a generic pagination response (deprecated, use specific types)
type PaginatorResponse struct {
	Pagination
	Result interface{} `json:"result"`
}

// NewPaginator creates a new Pagination response
func NewPaginator(c *echo.Context, result interface{}, total int64, page, limit int) *PaginatorResponse {
	return &PaginatorResponse{
		Pagination: NewPagination(total, page, limit),
		Result:     result,
	}
}

// NewPagination creates a new Pagination metadata
func NewPagination(total int64, page, limit int) Pagination {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	if page <= 0 {
		page = 1
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
	}
}

// GetPaginationParams extracts page and limit from the request query parameters
func GetPaginationParams(c *echo.Context) (page, limit int) {
	page, _ = strconv.Atoi(c.QueryParam("page"))
	limit, _ = strconv.Atoi(c.QueryParam("limit"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	return page, limit
}
