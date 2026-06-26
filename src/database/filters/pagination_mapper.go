package filters

import (
	"fmt"
	"strings"
)

// PaginationRequest represents a pagination request from the API
type PaginationRequest struct {
	Filter   string `json:"filter"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Sort     string `json:"sort"`
	Order    string `json:"order"`
}

// PaginationToQueryBuilder converts a PaginationRequest to a QueryBuilder
func PaginationToQueryBuilder(pr *PaginationRequest) *QueryBuilder {
	if pr == nil {
		return NewQueryBuilder("")
	}

	qb := NewQueryBuilder(pr.Filter)
	if qb.GetPagination() != nil {
		qb.GetPagination().Page = pr.Page
		qb.GetPagination().PageSize = pr.PageSize
	}

	if pr.Sort != "" {
		parts := strings.Split(pr.Sort, " ")
		field := parts[0]
		direction := "asc"
		if len(parts) > 1 {
			direction = parts[1]
		}
		if pr.Order != "" {
			direction = pr.Order
		}
		if qb.GetOrderBy() != nil {
			qb.GetOrderBy().Parse(fmt.Sprintf("%s %s", field, direction))
		}
	}

	return qb
}
