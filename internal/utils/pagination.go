package utils

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// PaginationQuery represents common query parameters for pagination, sorting and search
type PaginationQuery struct {
	Page    int    `form:"page" json:"page"`
	Limit   int    `form:"limit" json:"limit"`
	Search  string `form:"search" json:"search"`
	Sort    string `form:"sort" json:"sort"`
	SortBy  string `json:"-"`
	SortDir string `json:"-"`
}

// ParsePagination parses pagination, search and sort params from gin context
func ParsePagination(c *gin.Context) PaginationQuery {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	search := strings.TrimSpace(c.Query("search"))
	sort := strings.TrimSpace(c.Query("sort"))

	sortBy := "created_at"
	sortDir := "desc"

	if sort != "" {
		parts := strings.Split(sort, ":")
		if len(parts) >= 1 && parts[0] != "" {
			sortBy = parts[0]
		}
		if len(parts) >= 2 && strings.ToLower(parts[1]) == "asc" {
			sortDir = "asc"
		}
	}

	return PaginationQuery{
		Page:    page,
		Limit:   limit,
		Search:  search,
		Sort:    sort,
		SortBy:  sortBy,
		SortDir: sortDir,
	}
}

// GetOffset calculates offset for SQL query
func (p *PaginationQuery) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// BuildMeta generates PaginationMeta
func (p *PaginationQuery) BuildMeta(total int64) PaginationMeta {
	totalPages := 0
	if p.Limit > 0 {
		totalPages = int((total + int64(p.Limit) - 1) / int64(p.Limit))
	}
	return PaginationMeta{
		Page:       p.Page,
		Limit:      p.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
