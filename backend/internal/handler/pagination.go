package handler

import (
	"net/http"
	"strconv"

	"github.com/kunalchhabraiitb2024/taskflow/backend/internal/model"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// parsePagination reads ?page= and ?limit= query params with safe defaults.
func parsePagination(r *http.Request) model.PaginationParams {
	p := model.DefaultPagination()

	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			p.Page = n
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			if n > maxLimit {
				n = maxLimit
			}
			p.Limit = n
		}
	}
	return p
}
