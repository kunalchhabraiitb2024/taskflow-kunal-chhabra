package model

import "math"

// PaginationParams holds validated page/limit values.
type PaginationParams struct {
	Page  int // 1-based
	Limit int
}

// DefaultPagination returns sensible defaults when caller sends no params.
func DefaultPagination() PaginationParams {
	return PaginationParams{Page: 1, Limit: 20}
}

func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

// PaginatedResult wraps a slice of items with metadata for the client.
type PaginatedResult[T any] struct {
	Data       []T            `json:"data"`
	Pagination PaginationMeta `json:"pagination"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NewPaginatedResult[T any](items []T, total int, p PaginationParams) PaginatedResult[T] {
	if items == nil {
		items = []T{}
	}
	totalPages := int(math.Ceil(float64(total) / float64(p.Limit)))
	if totalPages == 0 {
		totalPages = 1
	}
	return PaginatedResult[T]{
		Data: items,
		Pagination: PaginationMeta{
			Page:       p.Page,
			Limit:      p.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	}
}
