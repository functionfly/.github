package pagination

import (
	"net/url"
	"strconv"

	"github.com/functionfly/functionfly/internal/apierror"
)

const (
	DefaultLimit  = 20
	MaxLimit     = 100
	DefaultCursor = ""
)

type Params struct {
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
	Cursor  string `json:"cursor,omitempty"`
	SortBy  string `json:"sort_by,omitempty"`
	SortDir string `json:"sort_dir,omitempty"`
}

type Response[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
	HasMore    bool  `json:"has_more"`
	NextCursor string `json:"next_cursor,omitempty"`
}

func NewResponse[T any](data []T, total int64, params Params) Response[T] {
	hasMore := int64(params.Offset+params.Limit) < total
	var nextCursor string
	if hasMore && params.Cursor == "" {
		nextCursor = EncodeCursor(params.Offset + params.Limit)
	}
	return Response[T]{
		Data:       data,
		Total:      total,
		Limit:      params.Limit,
		Offset:     params.Offset,
		HasMore:    hasMore,
		NextCursor: nextCursor,
	}
}

func EncodeCursor(offset int) string {
	return url.QueryEscape(strconv.Itoa(offset))
}

func DecodeCursor(cursor string) (int, error) {
	if cursor == "" {
		return 0, nil
	}
	decoded, err := url.QueryUnescape(cursor)
	if err != nil {
		return 0, apierror.NewInvalidCursor("Invalid cursor format")
	}
	offset, err := strconv.Atoi(decoded)
	if err != nil {
		return 0, apierror.NewInvalidCursor("Invalid cursor value")
	}
	if offset < 0 {
		return 0, apierror.NewInvalidCursor("Cursor cannot be negative")
	}
	return offset, nil
}

func ParseParams(r QueryParamsProvider) (Params, error) {
	params := Params{
		Limit: DefaultLimit,
	}

	if limitStr := r.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return params, apierror.NewInvalidLimit("limit must be a valid integer")
		}
		if limit < 0 {
			return params, apierror.NewInvalidLimit("limit cannot be negative")
		}
		if limit > MaxLimit {
			limit = MaxLimit
		}
		params.Limit = limit
	}

	if offsetStr := r.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return params, apierror.NewInvalidOffset("offset must be a valid integer")
		}
		if offset < 0 {
			return params, apierror.NewInvalidOffset("offset cannot be negative")
		}
		params.Offset = offset
	}

	if cursorStr := r.Get("cursor"); cursorStr != "" {
		offset, err := DecodeCursor(cursorStr)
		if err != nil {
			return params, err
		}
		params.Offset = offset
		params.Cursor = cursorStr
	}

	params.SortBy = r.Get("sort_by")
	params.SortDir = r.Get("sort_dir")

	return params, nil
}

type QueryParamsProvider interface {
	Get(key string) string
}

type PaginationConfig struct {
	DefaultLimit  int
	MaxLimit      int
	AllowedSortBy []string
}

func DefaultConfig() PaginationConfig {
	return PaginationConfig{
		DefaultLimit:  DefaultLimit,
		MaxLimit:      MaxLimit,
		AllowedSortBy:  []string{"created_at", "updated_at", "name", "id"},
	}
}

func (c PaginationConfig) ValidateSortBy(sortBy string) bool {
	if sortBy == "" {
		return true
	}
	for _, allowed := range c.AllowedSortBy {
		if sortBy == allowed {
			return true
		}
	}
	return false
}

func (c PaginationConfig) ValidateSortDir(sortDir string) bool {
	if sortDir == "" {
		return true
	}
	return sortDir == "asc" || sortDir == "desc"
}
