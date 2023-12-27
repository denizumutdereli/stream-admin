package database

import (
	"math"
)

type PaginatedResult struct {
	Data         interface{} `json:"data"`
	Total        int64       `json:"total"`
	CurrentPage  int         `json:"current_page"`
	NextPage     *int        `json:"next_page"`
	PreviousPage *int        `json:"previous_page"`
	PageSize     int         `json:"page_size"`
	TotalPages   int         `json:"total_pages"`
}

func PaginateTheResults(result interface{}, count int64, offset, page, limit int) *PaginatedResult {
	totalPages := int(math.Ceil(float64(count) / float64(limit)))

	nextPage := page + 1
	if nextPage > totalPages {
		nextPage = 0
	}

	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 0
	}

	paginatedResult := &PaginatedResult{
		Data:         result,
		Total:        count,
		CurrentPage:  page,
		NextPage:     &nextPage,
		PreviousPage: &prevPage,
		PageSize:     limit,
		TotalPages:   totalPages,
	}

	return paginatedResult
}