package filters

import (
	"errors"

	"gorm.io/gorm"
)

func PaginatedQuery[T any](
	db *gorm.DB,
	tenantID string,
	pagination *Pagination,
	model T,
	preloads ...string,
) (*PaginationResponse[T], error) {
	var items []T
	total := int64(0)
	if pagination == nil {
		pagination = &Pagination{
			Page:     1,
			PageSize: 10,
		}
	}

	// Get total count
	countQuery := db.Model(&model)
	if tenantID != "" {
		countQuery = countQuery.Where("tenant_id = ?", tenantID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}
	pagination.Total = total
	offset := pagination.GetOffset()

	query := db
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	for _, preload := range preloads {
		query = query.Preload(preload)
	}
	if err := query.Offset(offset).
		Limit(pagination.PageSize).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	response := PaginationResponse[T]{
		Items:      items,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: pagination.GetTotalPages(),
	}

	return &response, nil
}

func QueryDatabase[T any](
	db *gorm.DB,
	tenantID string,
	query_builder *QueryBuilder,
) (*QueryBuilderResponse[T], error) {
	var items []T
	var item T
	total := int64(0)
	if db == nil {
		return nil, errors.New("database is query is nil")
	}

	// applying the tenant_id filter
	if tenantID != "" {
		db = db.Where("tenant_id = ?", tenantID)
	}

	if query_builder == nil {
		query_builder = NewQueryBuilder("")
	}

	// Get total count for the database and query builder
	countQuery := db.Model(&item)
	if tenantID != "" {
		countQuery = countQuery.Where("tenant_id = ?", tenantID)
	}
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	query := query_builder.Apply(db)

	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}

	response := QueryBuilderResponse[T]{
		Items:      items,
		Total:      total,
		Page:       query_builder.GetPage(),
		PageSize:   query_builder.GetPageSize(),
		TotalPages: query_builder.GetTotalPages(),
	}

	return &response, nil
}
