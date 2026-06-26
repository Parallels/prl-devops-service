package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestModel represents a simple model for testing
type TestModel struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"column:name"`
	Age       int    `gorm:"column:age"`
	Status    string `gorm:"column:status"`
	CreatedAt int64  `gorm:"column:created_at"`
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&TestModel{})
	require.NoError(t, err)

	// Insert test data
	testData := []TestModel{
		{ID: 1, Name: "john", Age: 25, Status: "active", CreatedAt: 1000},
		{ID: 2, Name: "jane", Age: 30, Status: "inactive", CreatedAt: 2000},
		{ID: 3, Name: "bob", Age: 35, Status: "active", CreatedAt: 3000},
		{ID: 4, Name: "alice", Age: 28, Status: "active", CreatedAt: 4000},
		{ID: 5, Name: "charlie", Age: 32, Status: "inactive", CreatedAt: 5000},
	}

	for _, data := range testData {
		err = db.Create(&data).Error
		require.NoError(t, err)
	}

	return db
}

func TestNewQueryBuilder(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected struct {
			hasFilters    bool
			hasOrdering   bool
			hasPagination bool
		}
	}{
		{
			name: "empty string",
			raw:  "",
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
			}{false, true, true}, // default ordering + pagination defaults
		},
		{
			name: "full query string",
			raw:  "filter=name=john&order_by=created_at desc&page=1&page_size=10",
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
			}{true, true, true},
		},
		{
			name: "only filters",
			raw:  "name=john,age>25",
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
			}{true, true, true}, // parser interprets as both filter and ordering since = sign and field names
		},
		{
			name: "only ordering",
			raw:  "order_by=name asc,created_at desc",
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
			}{true, true, true}, // parser interprets as both filter and ordering
		},
		{
			name: "only pagination",
			raw:  "page=2&page_size=20",
			expected: struct {
				hasFilters    bool
				hasOrdering   bool
				hasPagination bool
			}{false, true, true}, // default ordering gets added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.raw)

			assert.NotNil(t, qb)
			assert.Equal(t, tt.expected.hasFilters, qb.HasFilters())
			assert.Equal(t, tt.expected.hasOrdering, qb.HasOrdering())
			assert.Equal(t, tt.expected.hasPagination, qb.HasPagination())
		})
	}
}

func TestQueryBuilder_GetComponents(t *testing.T) {
	qb := NewQueryBuilder("?filter=name=john&order_by=created_at desc&page=1&page_size=10")

	filter := qb.GetFilter()
	assert.NotNil(t, filter)
	assert.False(t, filter.IsEmpty())

	orderBy := qb.GetOrderBy()
	assert.NotNil(t, orderBy)
	assert.False(t, orderBy.IsEmpty())

	pagination := qb.GetPagination()
	assert.NotNil(t, pagination)
	assert.True(t, pagination.IsValid())
	assert.Equal(t, 1, pagination.GetPage())
	assert.Equal(t, 10, pagination.GetPageSize())
}

func TestQueryBuilder_Apply(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name           string
		query          string
		allowedColumns []string
		expectedCount  int64
		expectedIDs    []uint
	}{
		{
			name:           "filter only - status active",
			query:          "?filter=status=active",
			allowedColumns: []string{"status"},
			expectedCount:  3,
			expectedIDs:    []uint{1, 3, 4},
		},
		{
			name:           "filter and ordering",
			query:          "?filter=status=active&order_by=age desc",
			allowedColumns: []string{"status", "age"},
			expectedCount:  3,
			expectedIDs:    []uint{3, 4, 1}, // ordered by age desc: bob(35), alice(28), john(25)
		},
		{
			name:           "filter, ordering, and pagination",
			query:          "?filter=status=active&order_by=age desc&page=1&page_size=2",
			allowedColumns: []string{"status", "age"},
			expectedCount:  2,
			expectedIDs:    []uint{3, 4}, // first 2 results: bob(35), alice(28)
		},
		{
			name:           "pagination only",
			query:          "?page=2&page_size=2",
			allowedColumns: []string{},
			expectedCount:  2,
			expectedIDs:    []uint{3, 2}, // second page with default ordering "created_at desc": ID 3 (3000), ID 2 (2000)
		},
		{
			name:           "ordering only",
			query:          "?order_by=age desc",
			allowedColumns: []string{"age"},
			expectedCount:  5,
			expectedIDs:    []uint{3, 5, 2, 4, 1}, // ordered by age desc: bob(35), charlie(32), jane(30), alice(28), john(25)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.query)
			
			var results []TestModel
			query := qb.Apply(db.Model(&TestModel{}), tt.allowedColumns...)
			err := query.Find(&results).Error
			require.NoError(t, err)

			assert.Equal(t, tt.expectedCount, int64(len(results)))
			
			if len(tt.expectedIDs) > 0 {
				actualIDs := make([]uint, len(results))
				for i, result := range results {
					actualIDs[i] = result.ID
				}
				assert.Equal(t, tt.expectedIDs, actualIDs)
			}
		})
	}
}

func TestQueryBuilder_ApplyWithCount(t *testing.T) {
	db := setupTestDB(t)

	tests := []struct {
		name           string
		query          string
		allowedColumns []string
		expectedTotal  int64
		expectedCount  int64
		expectedPages  int
	}{
		{
			name:           "filter with pagination",
			query:          "?filter=status=active&page=1&page_size=2",
			allowedColumns: []string{"status"},
			expectedTotal:  3, // total active records
			expectedCount:  2, // first page results
			expectedPages:  2, // total pages
		},
		{
			name:           "no filters with pagination",
			query:          "?page=2&page_size=3",
			allowedColumns: []string{},
			expectedTotal:  5, // all records
			expectedCount:  2, // second page results (5 total, 3 per page = 2 on page 2)
			expectedPages:  2, // total pages
		},
		{
			name:           "filter that returns no results",
			query:          "?filter=status=nonexistent&page=1&page_size=10",
			allowedColumns: []string{"status"},
			expectedTotal:  0,
			expectedCount:  0,
			expectedPages:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.query)
			
			var results []TestModel
			query, err := qb.ApplyWithCount(db.Model(&TestModel{}), tt.allowedColumns...)
			require.NoError(t, err)
			
			err = query.Find(&results).Error
			require.NoError(t, err)

			assert.Equal(t, tt.expectedTotal, qb.GetPagination().GetTotal())
			assert.Equal(t, tt.expectedCount, int64(len(results)))
			assert.Equal(t, tt.expectedPages, qb.GetPagination().GetTotalPages())
		})
	}
}

func TestQueryBuilder_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected bool
	}{
		{
			name:     "empty query",
			query:    "",
			expected: false, // pagination has default values, so not empty
		},
		{
			name:     "query with filters",
			query:    "name=john",
			expected: false,
		},
		{
			name:     "query with ordering",
			query:    "order_by=name asc",
			expected: false,
		},
		{
			name:     "query with pagination",
			query:    "page=1&page_size=10",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.query)
			assert.Equal(t, tt.expected, qb.IsEmpty())
		})
	}
}

func TestQueryBuilder_ComponentChecks(t *testing.T) {
	qb := NewQueryBuilder("?filter=name=john&order_by=created_at desc&page=1&page_size=10")

	assert.True(t, qb.HasFilters())
	assert.True(t, qb.HasOrdering())
	assert.True(t, qb.HasPagination())

	// Test with truly empty components (but note: ordering now has a default)
	emptyQb := NewQueryBuilder("")
	assert.False(t, emptyQb.HasFilters())
	assert.True(t, emptyQb.HasOrdering()) // now has default ordering: created_at desc
	assert.True(t, emptyQb.HasPagination()) // pagination defaults to valid values
}

func TestQueryBuilder_NilDatabase(t *testing.T) {
	qb := NewQueryBuilder("?name=john&order_by=created_at desc&page=1&page_size=10")
	
	// Test Apply with nil database
	result := qb.Apply(nil)
	assert.Nil(t, result)
	
	// Test ApplyWithCount with nil database
	result, err := qb.ApplyWithCount(nil)
	assert.Nil(t, result)
	assert.NoError(t, err)
}

func TestQueryBuilder_AllowedColumns(t *testing.T) {
	db := setupTestDB(t)
	
	// Test with restricted columns
	qb := NewQueryBuilder("?filter=name=john,age>20&order_by=status asc")
	
	var results []TestModel
	// Only allow 'name' column, should ignore 'age' filter and 'status' ordering
	query := qb.Apply(db.Model(&TestModel{}), "name")
	err := query.Find(&results).Error
	require.NoError(t, err)
	
	// Should only apply name filter, not age filter or status ordering
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "john", results[0].Name)
}

func TestQueryBuilder_ComplexQueries(t *testing.T) {
	db := setupTestDB(t)
	
	tests := []struct {
		name          string
		query         string
		expectedCount int
		description   string
	}{
		{
			name:          "multiple filters with AND",
			query:         "?filter=age>25 AND status=active&order_by=created_at desc&page=1&page_size=10",
			expectedCount: 2,
			description:   "should find bob and alice (age > 25 AND status = active)",
		},
		{
			name:          "filter with ordering and pagination",
			query:         "?filter=status=active&order_by=-created_at&page=1&page_size=2",
			expectedCount: 2,
			description:   "should find first 2 active users ordered by created_at desc",
		},
		{
			name:          "ordering only",
			query:         "?order_by=age asc",
			expectedCount: 5,
			description:   "should find all users ordered by age ascending",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.query)
			
			var results []TestModel
			query := qb.Apply(db.Model(&TestModel{}), "age", "status", "created_at")
			err := query.Find(&results).Error
			require.NoError(t, err)
			
			assert.Equal(t, tt.expectedCount, len(results), tt.description)
		})
	}
}

func TestQueryBuilder_PaginationHelpers(t *testing.T) {
	tests := []struct {
		name                string
		query               string
		total               int64
		expectedPage        int
		expectedPageSize    int
		expectedTotalPages  int
		expectedOffset      int
	}{
		{
			name:               "default pagination",
			query:              "",
			total:              0,
			expectedPage:       1,
			expectedPageSize:   20, // default from config (DefaultPageSizeInt = 20)
			expectedTotalPages: 0,
			expectedOffset:     0,
		},
		{
			name:               "custom pagination",
			query:              "?page=2&page_size=5",
			total:              23,
			expectedPage:       2,
			expectedPageSize:   5,
			expectedTotalPages: 5, // ceil(23/5) = 5
			expectedOffset:     5, // (2-1) * 5 = 5
		},
		{
			name:               "first page",
			query:              "?page=1&page_size=10",
			total:              50,
			expectedPage:       1,
			expectedPageSize:   10,
			expectedTotalPages: 5, // ceil(50/10) = 5
			expectedOffset:     0, // (1-1) * 10 = 0
		},
		{
			name:               "last page partial",
			query:              "?page=3&page_size=10",
			total:              25,
			expectedPage:       3,
			expectedPageSize:   10,
			expectedTotalPages: 3, // ceil(25/10) = 3
			expectedOffset:     20, // (3-1) * 10 = 20
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.query)
			
			// Set total if provided
			if tt.total > 0 {
				qb.SetTotalPages(tt.total)
			}
			
			assert.Equal(t, tt.expectedPage, qb.GetPage())
			assert.Equal(t, tt.expectedPageSize, qb.GetPageSize())
			assert.Equal(t, tt.expectedTotalPages, qb.GetTotalPages())
			assert.Equal(t, tt.expectedOffset, qb.GetOffset())
		})
	}
}

func TestQueryBuilder_SetTotalPages(t *testing.T) {
	qb := NewQueryBuilder("?page=2&page_size=5")
	
	// Initially no total set
	assert.Equal(t, 0, qb.GetTotalPages())
	
	// Set total
	qb.SetTotalPages(23)
	assert.Equal(t, 5, qb.GetTotalPages()) // ceil(23/5) = 5
	
	// Change total
	qb.SetTotalPages(10)
	assert.Equal(t, 2, qb.GetTotalPages()) // ceil(10/5) = 2
	
	// Set zero total
	qb.SetTotalPages(0)
	assert.Equal(t, 0, qb.GetTotalPages())
}

func TestQueryBuilder_SetTotalPagesWithNilPagination(t *testing.T) {
	qb := &QueryBuilder{} // No pagination initialized
	
	// Should not panic
	qb.SetTotalPages(100)
	
	// Should return 0 since pagination is nil
	assert.Equal(t, 0, qb.GetTotalPages())
}

func TestQueryBuilder_GettersWithNilPagination(t *testing.T) {
	qb := &QueryBuilder{} // No pagination initialized
	
	assert.Equal(t, 0, qb.GetPage())
	assert.Equal(t, -1, qb.GetPageSize())
	assert.Equal(t, 0, qb.GetTotalPages())
	assert.Equal(t, 0, qb.GetOffset())
}

func TestQueryBuilder_DefaultOrderBy(t *testing.T) {
	tests := []struct {
		name               string
		query              string
		expectedHasOrdering bool
		expectedOrderClause string
	}{
		{
			name:               "empty query gets default ordering",
			query:              "",
			expectedHasOrdering: true,
			expectedOrderClause: "created_at desc",
		},
		{
			name:               "query with explicit ordering keeps it",
			query:              "?order_by=name asc",
			expectedHasOrdering: true,
			expectedOrderClause: "name asc",
		},
		{
			name:               "query with filter but no ordering gets default",
			query:              "?filter=name=john",
			expectedHasOrdering: true,
			expectedOrderClause: "created_at desc",
		},
		{
			name:               "pagination only gets default ordering",
			query:              "?page=2&page_size=10",
			expectedHasOrdering: true,
			expectedOrderClause: "created_at desc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.query)
			
			assert.Equal(t, tt.expectedHasOrdering, qb.HasOrdering())
			
			// Check that the ordering clause is as expected
			clauses := qb.GetOrderBy().GetClauses()
			if tt.expectedHasOrdering && len(clauses) > 0 {
				// For default ordering, should be created_at desc
				if tt.expectedOrderClause == "created_at desc" {
					assert.Equal(t, "created_at", clauses[0].Column)
					assert.Equal(t, SortDesc, clauses[0].Direction)
				}
			}
		})
	}
}

func TestQueryBuilder_PaginationEdgeCases(t *testing.T) {
	tests := []struct {
		name             string
		query            string
		total            int64
		expectedOffset   int
		expectedPage     int
		description      string
	}{
		{
			name:           "page beyond total records",
			query:          "?page=10&page_size=5",
			total:          12, // only 3 pages worth
			expectedOffset: 7,  // Total(12) - PageSize(5) = 7 (adjusted to show last data)
			expectedPage:   1,  // Page gets reset to 1 when beyond available data
			description:    "should handle pages beyond available data by adjusting to last valid page",
		},
		{
			name:           "very large page size",
			query:          "?page=1&page_size=1000",
			total:          50,
			expectedOffset: 0,
			expectedPage:   1,
			description:    "should handle page size larger than total records",
		},
		{
			name:           "total equals page size",
			query:          "?page=1&page_size=10",
			total:          10,
			expectedOffset: 0,
			expectedPage:   1,
			description:    "should handle exact page size match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder(tt.query)
			qb.SetTotalPages(tt.total)
			
			assert.Equal(t, tt.expectedOffset, qb.GetOffset(), tt.description)
			assert.Equal(t, tt.expectedPage, qb.GetPage(), tt.description)
		})
	}
}