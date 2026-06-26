package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewPaginationFilter(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected *PaginationFilter
	}{
		{
			name: "empty string",
			raw:  "",
			expected: &PaginationFilter{
				Page:     1,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name: "simple pagination parameters",
			raw:  "page=2&page_size=20",
			expected: &PaginationFilter{
				Page:     2,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name: "pagination with query string prefix",
			raw:  "?page=3&page_size=15",
			expected: &PaginationFilter{
				Page:     3,
				PageSize: 15,
				Total:    0,
			},
		},
		{
			name: "full query string with other parameters",
			raw:  "filter=active&page=5&page_size=25&order_by=name",
			expected: &PaginationFilter{
				Page:     5,
				PageSize: 25,
				Total:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewPaginationFilter(tt.raw)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationFilter_Parse(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected *PaginationFilter
	}{
		{
			name: "empty string",
			raw:  "",
			expected: &PaginationFilter{
				Page:     1,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name: "whitespace only",
			raw:  "   ",
			expected: &PaginationFilter{
				Page:     1,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name: "basic page and page_size",
			raw:  "page=2&page_size=20",
			expected: &PaginationFilter{
				Page:     2,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name: "with query prefix",
			raw:  "?page=3&page_size=15",
			expected: &PaginationFilter{
				Page:     3,
				PageSize: 15,
				Total:    0,
			},
		},
		{
			name: "alternative page size parameters",
			raw:  "page=1&limit=50",
			expected: &PaginationFilter{
				Page:     1,
				PageSize: 50,
				Total:    0,
			},
		},
		{
			name: "pageSize parameter",
			raw:  "page=2&pageSize=30",
			expected: &PaginationFilter{
				Page:     2,
				PageSize: 30,
				Total:    0,
			},
		},
		{
			name: "per_page parameter",
			raw:  "page=4&per_page=40",
			expected: &PaginationFilter{
				Page:     4,
				PageSize: 40,
				Total:    0,
			},
		},
		{
			name: "perpage parameter",
			raw:  "page=5&perpage=35",
			expected: &PaginationFilter{
				Page:     5,
				PageSize: 35,
				Total:    0,
			},
		},
		{
			name: "case insensitive parameters",
			raw:  "PAGE=6&PAGE_SIZE=45",
			expected: &PaginationFilter{
				Page:     6,
				PageSize: 45,
				Total:    0,
			},
		},
		{
			name: "invalid page number",
			raw:  "page=0&page_size=20",
			expected: &PaginationFilter{
				Page:     1,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name: "negative page number",
			raw:  "page=-1&page_size=20",
			expected: &PaginationFilter{
				Page:     1,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name: "invalid page size",
			raw:  "page=2&page_size=0",
			expected: &PaginationFilter{
				Page:     2,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name: "negative page size",
			raw:  "page=2&page_size=-10",
			expected: &PaginationFilter{
				Page:     2,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name: "non-numeric values",
			raw:  "page=abc&page_size=def",
			expected: &PaginationFilter{
				Page:     1,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name: "mixed valid and invalid parameters",
			raw:  "page=3&invalid=value&page_size=25&another=param",
			expected: &PaginationFilter{
				Page:     3,
				PageSize: 25,
				Total:    0,
			},
		},
		{
			name: "parameters without values",
			raw:  "page&page_size=20",
			expected: &PaginationFilter{
				Page:     1,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name: "extra whitespace",
			raw:  " page = 2 & page_size = 20 ",
			expected: &PaginationFilter{
				Page:     2,
				PageSize: 20,
				Total:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := &PaginationFilter{
				Page:     1,
				PageSize: 20,
				Total:    0,
			}
			pf.Parse(tt.raw)
			assert.Equal(t, tt.expected, pf)
		})
	}
}

func TestPaginationFilter_GetPage(t *testing.T) {
	pf := &PaginationFilter{Page: 5}
	assert.Equal(t, 5, pf.GetPage())
}

func TestPaginationFilter_GetPageSize(t *testing.T) {
	pf := &PaginationFilter{PageSize: 25}
	assert.Equal(t, 25, pf.GetPageSize())
}

func TestPaginationFilter_GetTotal(t *testing.T) {
	pf := &PaginationFilter{Total: 100}
	assert.Equal(t, int64(100), pf.GetTotal())
}

func TestPaginationFilter_SetTotal(t *testing.T) {
	pf := &PaginationFilter{}
	pf.SetTotal(150)
	assert.Equal(t, int64(150), pf.Total)
	assert.Equal(t, int64(150), pf.GetTotal())
}

func TestPaginationFilter_GetTotalPages(t *testing.T) {
	tests := []struct {
		name     string
		pf       *PaginationFilter
		expected int
	}{
		{
			name: "zero total",
			pf: &PaginationFilter{
				Total:    0,
				PageSize: 10,
			},
			expected: 0,
		},
		{
			name: "zero page size",
			pf: &PaginationFilter{
				Total:    100,
				PageSize: 0,
			},
			expected: 0,
		},
		{
			name: "exact division",
			pf: &PaginationFilter{
				Total:    100,
				PageSize: 10,
			},
			expected: 10,
		},
		{
			name: "with remainder",
			pf: &PaginationFilter{
				Total:    103,
				PageSize: 10,
			},
			expected: 11,
		},
		{
			name: "single page",
			pf: &PaginationFilter{
				Total:    5,
				PageSize: 10,
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pf.GetTotalPages()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationFilter_GetOffset(t *testing.T) {
	tests := []struct {
		name     string
		pf       *PaginationFilter
		expected int
	}{
		{
			name: "zero total",
			pf: &PaginationFilter{
				Page:     2,
				PageSize: 10,
				Total:    0,
			},
			expected: 10,
		},
		{
			name: "first page",
			pf: &PaginationFilter{
				Page:     1,
				PageSize: 10,
				Total:    100,
			},
			expected: 0,
		},
		{
			name: "second page",
			pf: &PaginationFilter{
				Page:     2,
				PageSize: 10,
				Total:    100,
			},
			expected: 10,
		},
		{
			name: "third page",
			pf: &PaginationFilter{
				Page:     3,
				PageSize: 15,
				Total:    100,
			},
			expected: 30,
		},
		{
			name: "page beyond total resets to page 1",
			pf: &PaginationFilter{
				Page:     20,
				PageSize: 10,
				Total:    50,
			},
			expected: 40,
		},
		{
			name: "page size larger than total",
			pf: &PaginationFilter{
				Page:     2,
				PageSize: 100,
				Total:    50,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pf.GetOffset()
			assert.Equal(t, tt.expected, result)

			// Check if page was reset when offset was beyond total
			if tt.name == "page beyond total resets to page 1" {
				assert.Equal(t, 1, tt.pf.Page)
			}
		})
	}
}

func TestPaginationFilter_GetPageIndex(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		expected int
	}{
		{"page 1", 1, 0},
		{"page 2", 2, 1},
		{"page 5", 5, 4},
		{"page 10", 10, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := &PaginationFilter{Page: tt.page}
			result := pf.GetPageIndex()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationFilter_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		pf       *PaginationFilter
		expected bool
	}{
		{
			name: "valid pagination",
			pf: &PaginationFilter{
				Page:     1,
				PageSize: 10,
			},
			expected: true,
		},
		{
			name: "zero page",
			pf: &PaginationFilter{
				Page:     0,
				PageSize: 10,
			},
			expected: false,
		},
		{
			name: "negative page",
			pf: &PaginationFilter{
				Page:     -1,
				PageSize: 10,
			},
			expected: false,
		},
		{
			name: "zero page size",
			pf: &PaginationFilter{
				Page:     1,
				PageSize: 0,
			},
			expected: false,
		},
		{
			name: "negative page size",
			pf: &PaginationFilter{
				Page:     1,
				PageSize: -10,
			},
			expected: false,
		},
		{
			name: "both invalid",
			pf: &PaginationFilter{
				Page:     0,
				PageSize: 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pf.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPaginationFilter_Apply(t *testing.T) {
	// Setup in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	tests := []struct {
		name string
		pf   *PaginationFilter
		db   *gorm.DB
	}{
		{
			name: "nil database",
			pf: &PaginationFilter{
				Page:     1,
				PageSize: 10,
				Total:    100,
			},
			db: nil,
		},
		{
			name: "invalid pagination",
			pf: &PaginationFilter{
				Page:     0,
				PageSize: 10,
				Total:    100,
			},
			db: db,
		},
		{
			name: "valid pagination",
			pf: &PaginationFilter{
				Page:     2,
				PageSize: 15,
				Total:    100,
			},
			db: db,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pf.Apply(tt.db)

			if tt.name == "nil database" {
				assert.Nil(t, result)
			} else if tt.name == "invalid pagination" {
				assert.Equal(t, tt.db, result)
			} else {
				assert.NotNil(t, result)
				// The result should be a different DB instance with applied pagination
				assert.NotEqual(t, tt.db, result)
			}
		})
	}
}

func TestPaginationFilter_ExtractPaginationFromQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected *PaginationFilter
	}{
		{
			name:  "empty query",
			query: "",
			expected: &PaginationFilter{
				Page:     1,
				PageSize: 10,
				Total:    0,
			},
		},
		{
			name:  "query with leading question mark",
			query: "?page=3&page_size=20",
			expected: &PaginationFilter{
				Page:     3,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name:  "query without leading question mark",
			query: "page=4&page_size=25",
			expected: &PaginationFilter{
				Page:     4,
				PageSize: 25,
				Total:    0,
			},
		},
		{
			name:  "complex query with other parameters",
			query: "filter=active&sort=name&page=2&page_size=30&order=asc",
			expected: &PaginationFilter{
				Page:     2,
				PageSize: 30,
				Total:    0,
			},
		},
		{
			name:  "query with empty parameters",
			query: "page=&page_size=20",
			expected: &PaginationFilter{
				Page:     1,
				PageSize: 20,
				Total:    0,
			},
		},
		{
			name:  "query with malformed parameters",
			query: "page&page_size=20&invalid",
			expected: &PaginationFilter{
				Page:     1,
				PageSize: 20,
				Total:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pf := &PaginationFilter{
				Page:     1,
				PageSize: 10,
				Total:    0,
			}
			pf.extractPaginationFromQuery(tt.query)
			assert.Equal(t, tt.expected, pf)
		})
	}
}
