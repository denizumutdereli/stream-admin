package middleware

import (
	"fmt"
	"strconv"

	"github.com/denizumutdereli/stream-admin/internal/types"
	"github.com/gin-gonic/gin"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 10
)

func Pagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
		if err != nil || page < 1 {
			page = DefaultPage
		}

		pageSize, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultPageSize)))
		if err != nil || pageSize < 1 {
			pageSize = DefaultPageSize
		}

		sortBy := c.DefaultQuery("sortBy", "")
		sortOrder := c.DefaultQuery("sortOrder", "")

		paginationParams := types.PaginationParams{
			Page:      page,
			Limit:     pageSize,
			SortBy:    sortBy,
			SortOrder: sortOrder,
		}

		fmt.Println(paginationParams)
		c.Set("pagination", paginationParams)

		c.Next()
	}
}

func SanitizePaginationParams(params *types.PaginationParams) *types.PaginationParams {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = DefaultPageSize
	}
	if params.SortBy == "" {
		params.SortBy = "id"
	}
	if params.SortOrder != "asc" && params.SortOrder != "desc" {
		params.SortOrder = "asc"
	}
	return params
}
