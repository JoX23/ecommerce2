package domain

// PaginationParams holds page/limit values for paginated queries.
type PaginationParams struct {
	Page  int // 1-based
	Limit int
}

// Offset returns the SQL/slice offset for the current page.
func (p PaginationParams) Offset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.Limit
}

// DefaultPagination returns safe defaults: page 1, 20 items.
func DefaultPagination() PaginationParams {
	return PaginationParams{Page: 1, Limit: 20}
}

// PaginatedResult wraps a slice of results with pagination metadata.
type PaginatedResult[T any] struct {
	Data       []T `json:"data"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}
