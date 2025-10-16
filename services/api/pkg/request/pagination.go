package request

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Pagination defines pagination query parameters for the API.
// @Param page query int false "Page number, starting from 1" minimum(1)
// @Param size query int false "Number of items per page" minimum(1) maximum(100) name(size)
// @Param order query string false "Field to order by"
// @Param sort query string false "Sort order, either ASC or DESC" Enums(ASC, DESC)
// @Description Pagination parameters
type Pagination struct {
	Page     int    `json:"page" query:"page" validate:"omitempty,min=1"`
	PageSize int    `json:"size" query:"size" validate:"omitempty,min=1,max=100"`
	Order    string `json:"order" query:"order" validate:"omitempty"`
	Sort     string `json:"sort" query:"sort" validate:"omitempty,oneof=DESC ASC"`
}

func (r *Request) Pagination(c echo.Context) *Pagination {
	var pagination Pagination

	if err := c.Bind(&pagination); err != nil {
		return &Pagination{
			Page:     1,
			PageSize: 50,
			Order:    "id",
			Sort:     "ASC",
		}
	}

	if pagination.Page == 0 {
		pagination.Page = 1
	}
	if pagination.PageSize == 0 {
		pagination.PageSize = 50
	}
	if pagination.Order == "" {
		pagination.Order = "id"
	}
	if pagination.Sort == "" {
		pagination.Sort = "ASC"
	}

	return &pagination
}

type Meta struct {
	Page         int   `json:"page" validate:"required"`
	PageSize     int   `json:"size" validate:"required"`
	TotalRecords int64 `json:"total_records" validate:"required"`
}

func Paginator(ctx context.Context, db *gorm.DB, pagination *Pagination) *gorm.DB {
	if pagination.Order == "" {
		pagination.Order = "id"
		pagination.Sort = "DESC"
	}

	offset := (pagination.Page - 1) * pagination.PageSize

	order := fmt.Sprintf("%s %s", pagination.Order, pagination.Sort)

	return db.WithContext(ctx).Order(order).Limit(pagination.PageSize).Offset(offset)
}
