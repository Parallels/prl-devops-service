package filters

import (
	"gorm.io/gorm"
)

// QueryBuilder combines filtering, ordering, and pagination into a single unified interface
type QueryBuilder struct {
	filter     *FilterQuery
	orderBy    *OrderByFilter
	pagination *PaginationFilter
}

// NewQueryBuilder creates a new QueryBuilder instance and parses the raw query string
// Supports various query string formats:
// - "filter=name=john&order_by=created_at desc&page=1&page_size=10"
// - "name=john,age>25&sort=name asc&limit=20"
// - Simple filter string: "name=john AND age>25"
func NewQueryBuilder(raw string) *QueryBuilder {
	builder := &QueryBuilder{
		filter:     NewFilterQuery(raw),
		orderBy:    NewOrderByFilter(raw),
		pagination: NewPaginationFilter(raw),
	}

	// Set default order by if not specified
	if builder.orderBy.IsEmpty() {
		builder.orderBy = NewOrderByFilter("created_at desc")
	}
	return builder
}

// GetFilter returns the FilterQuery instance
func (qb *QueryBuilder) GetFilter() *FilterQuery {
	return qb.filter
}

// GetOrderBy returns the OrderByFilter instance
func (qb *QueryBuilder) GetOrderBy() *OrderByFilter {
	return qb.orderBy
}

// GetPagination returns the PaginationFilter instance
func (qb *QueryBuilder) GetPagination() *PaginationFilter {
	return qb.pagination
}

// SetTotalPages sets the total count of records for pagination calculations
// This is typically called after executing a count query to provide pagination metadata
func (qb *QueryBuilder) SetTotalPages(total int64) {
	if qb.pagination != nil {
		qb.pagination.Total = total
	}
}

// GetTotalPages returns the total number of pages based on total records and page size
// Returns 0 if pagination is not configured or total is not set
func (qb *QueryBuilder) GetTotalPages() int {
	if qb.pagination != nil {
		return qb.pagination.GetTotalPages()
	}
	return 0
}

// GetOffset returns the database offset for the current page
// Used for LIMIT/OFFSET database queries
// Returns 0 if pagination is not configured
func (qb *QueryBuilder) GetOffset() int {
	if qb.pagination != nil {
		return qb.pagination.GetOffset()
	}
	return 0
}

// GetPageSize returns the number of records per page
// Returns -1 if pagination is not configured, indicating no pagination limit
func (qb *QueryBuilder) GetPageSize() int {
	if qb.pagination != nil {
		return qb.pagination.PageSize
	}
	return -1
}

// GetPage returns the current page number (1-based)
// Returns 0 if pagination is not configured
func (qb *QueryBuilder) GetPage() int {
	if qb.pagination != nil {
		return qb.pagination.Page
	}
	return 0
}

// Apply applies all filters, ordering, and pagination to a gorm DB and returns the updated DB
// The order of application is: filters -> ordering -> pagination
// This ensures correct SQL generation and optimal query performance
func (qb *QueryBuilder) Apply(db *gorm.DB, allowedColumns ...string) *gorm.DB {
	if db == nil {
		return db
	}

	// Apply filters first
	if qb.filter != nil && !qb.filter.IsEmpty() {
		db = qb.filter.Apply(db, allowedColumns...)
	}

	// Apply ordering second
	if qb.orderBy != nil && !qb.orderBy.IsEmpty() {
		db = qb.orderBy.Apply(db, allowedColumns...)
	}

	// Apply pagination last
	if qb.pagination != nil && qb.pagination.IsValid() {
		db = qb.pagination.Apply(db)
	}

	return db
}

// ApplyWithCount applies filters and ordering, gets the total count, then applies pagination
// This is useful when you need to know the total number of records for pagination metadata
// Returns the modified DB and sets the total count in the pagination filter
func (qb *QueryBuilder) ApplyWithCount(db *gorm.DB, allowedColumns ...string) (*gorm.DB, error) {
	if db == nil {
		return db, nil
	}

	// Apply filters first
	if qb.filter != nil && !qb.filter.IsEmpty() {
		db = qb.filter.Apply(db, allowedColumns...)
	}

	// Get total count before applying pagination
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return db, err
	}

	// Set total in pagination
	if qb.pagination != nil {
		qb.pagination.SetTotal(total)
	}

	// Apply ordering
	if qb.orderBy != nil && !qb.orderBy.IsEmpty() {
		db = qb.orderBy.Apply(db, allowedColumns...)
	}

	// Apply pagination last
	if qb.pagination != nil && qb.pagination.IsValid() {
		db = qb.pagination.Apply(db)
	}

	return db, nil
}

// IsEmpty returns true if all components (filter, orderBy, pagination) are empty or invalid
func (qb *QueryBuilder) IsEmpty() bool {
	filterEmpty := qb.filter == nil || qb.filter.IsEmpty()
	orderByEmpty := qb.orderBy == nil || qb.orderBy.IsEmpty()
	paginationInvalid := qb.pagination == nil || !qb.pagination.IsValid()

	return filterEmpty && orderByEmpty && paginationInvalid
}

// HasFilters returns true if there are filter conditions
func (qb *QueryBuilder) HasFilters() bool {
	return qb.filter != nil && !qb.filter.IsEmpty()
}

// HasOrdering returns true if there are ordering clauses
func (qb *QueryBuilder) HasOrdering() bool {
	return qb.orderBy != nil && !qb.orderBy.IsEmpty()
}

// HasPagination returns true if pagination is valid
func (qb *QueryBuilder) HasPagination() bool {
	return qb.pagination != nil && qb.pagination.IsValid()
}
