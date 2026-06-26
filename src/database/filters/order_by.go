package filters

import (
	"strings"

	"gorm.io/gorm"
)

// SortDirection represents the SQL sort direction
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// OrderByClause represents a single ORDER BY expression
type OrderByClause struct {
	Column    string
	Direction SortDirection
}

// OrderByFilter represents a collection of ORDER BY clauses for URL query parsing
type OrderByFilter struct {
	clauses []OrderByClause
}

// NewOrderByFilter creates a new OrderByFilter instance and parses the raw string
func NewOrderByFilter(raw string) *OrderByFilter {
	ob := &OrderByFilter{
		clauses: make([]OrderByClause, 0),
	}
	if raw != "" {
		ob.Parse(raw)
	}

	return ob
}

// Parse parses a raw order_by string (e.g. "name asc,created_at desc" or "name:asc,created_at:desc")
// into OrderByClause slices. Unspecified directions default to asc.
// If a full query string is provided (e.g. "filter=active&page=1&page_size=10&order_by=name asc"),
// it will extract only the order_by parameter and ignore the rest.
func (ob *OrderByFilter) Parse(raw string) {
	ob.clauses = make([]OrderByClause, 0)
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return
	}

	// Check if this looks like a full query string with multiple parameters
	if strings.Contains(trimmed, "&") || strings.Contains(trimmed, "?") {
		// Extract only the order_by parameter
		orderByValue := ob.extractOrderByFromQuery(trimmed)
		if orderByValue == "" {
			return
		}
		trimmed = orderByValue
	}

	// split by comma to support multiple fields
	parts := strings.Split(trimmed, ",")
	for _, part := range parts {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}

		field := p
		dir := SortAsc

		// support delimiters: space or colon
		if strings.Contains(p, " ") {
			sp := strings.Fields(p)
			field = sp[0]
			if len(sp) > 1 {
				d := strings.ToLower(sp[1])
				if d == string(SortDesc) {
					dir = SortDesc
				}
			}
		} else if strings.Contains(p, ":") {
			sp := strings.SplitN(p, ":", 2)
			field = strings.TrimSpace(sp[0])
			if len(sp) > 1 {
				d := strings.ToLower(strings.TrimSpace(sp[1]))
				if d == string(SortDesc) {
					dir = SortDesc
				}
			}
		}

		// prefix "-" indicates desc (e.g. "-created_at")
		if strings.HasPrefix(field, "-") {
			field = strings.TrimPrefix(field, "-")
			dir = SortDesc
		}

		if field == "" {
			continue
		}

		ob.clauses = append(ob.clauses, OrderByClause{Column: field, Direction: dir})
	}
}

// extractOrderByFromQuery extracts the order_by parameter value from a full query string
func (ob *OrderByFilter) extractOrderByFromQuery(query string) string {
	// Remove leading ? if present
	query = strings.TrimPrefix(query, "?")

	// Split by & to get individual parameters
	params := strings.Split(query, "&")

	for _, param := range params {
		param = strings.TrimSpace(param)
		if param == "" {
			continue
		}

		// Split by = to get key-value pairs
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Check for order_by parameter (case insensitive)
		if strings.EqualFold(key, "order_by") || strings.EqualFold(key, "orderby") || strings.EqualFold(key, "sort") {
			return value
		}
	}

	return ""
}

// GetClauses returns the parsed OrderByClause slices
func (ob *OrderByFilter) GetClauses() []OrderByClause {
	return ob.clauses
}

// IsEmpty returns true if no clauses were parsed
func (ob *OrderByFilter) IsEmpty() bool {
	return len(ob.clauses) == 0
}

// Apply applies the order by clauses to a gorm DB and returns the updated DB
// Example accepted raw formats:
// - "name asc, created_at desc"
// - "name:asc,created_at:desc"
// - "name, -created_at"
func (ob *OrderByFilter) Apply(db *gorm.DB, allowedColumns ...string) *gorm.DB {
	if db == nil || ob.IsEmpty() {
		return db
	}

	// Optional allow-list to mitigate SQL injection via column names
	isAllowed := func(col string) bool { return true }
	if len(allowedColumns) > 0 {
		allowed := make(map[string]struct{}, len(allowedColumns))
		for _, c := range allowedColumns {
			allowed[strings.ToLower(strings.TrimSpace(c))] = struct{}{}
		}
		isAllowed = func(col string) bool {
			_, ok := allowed[strings.ToLower(strings.TrimSpace(col))]
			return ok
		}
	}

	for _, c := range ob.clauses {
		if c.Column == "" || !isAllowed(c.Column) {
			continue
		}
		dir := "ASC"
		if c.Direction == SortDesc {
			dir = "DESC"
		}
		db = db.Order(c.Column + " " + dir)
	}
	return db
}
