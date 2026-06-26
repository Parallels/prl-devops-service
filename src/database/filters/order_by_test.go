package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestOrderByModel represents a simple model for testing ordering
type TestOrderByModel struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"column:name"`
	Age       int    `gorm:"column:age"`
	CreatedAt int64  `gorm:"column:created_at"`
	UpdatedAt int64  `gorm:"column:updated_at"`
}

func setupOrderByTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&TestOrderByModel{})
	require.NoError(t, err)

	// Insert test data in mixed order to test sorting
	testData := []TestOrderByModel{
		{ID: 1, Name: "charlie", Age: 35, CreatedAt: 3000, UpdatedAt: 1000},
		{ID: 2, Name: "alice", Age: 25, CreatedAt: 1000, UpdatedAt: 3000},
		{ID: 3, Name: "bob", Age: 30, CreatedAt: 2000, UpdatedAt: 2000},
		{ID: 4, Name: "diana", Age: 28, CreatedAt: 4000, UpdatedAt: 4000},
	}

	for _, data := range testData {
		err = db.Create(&data).Error
		require.NoError(t, err)
	}

	return db
}

func TestNewOrderByFilter(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected int // number of expected clauses
	}{
		{
			name:     "empty string",
			raw:      "",
			expected: 0,
		},
		{
			name:     "single field default direction",
			raw:      "name",
			expected: 1,
		},
		{
			name:     "single field with space delimiter",
			raw:      "name asc",
			expected: 1,
		},
		{
			name:     "single field with colon delimiter",
			raw:      "name:desc",
			expected: 1,
		},
		{
			name:     "single field with minus prefix",
			raw:      "-created_at",
			expected: 1,
		},
		{
			name:     "multiple fields comma separated",
			raw:      "name asc,created_at desc",
			expected: 2,
		},
		{
			name:     "query string format",
			raw:      "?order_by=name asc&page=1",
			expected: 1,
		},
		{
			name:     "complex query string",
			raw:      "filter=active&order_by=name desc,created_at asc&page_size=10",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewOrderByFilter(tt.raw)
			assert.Equal(t, tt.expected, len(result.GetClauses()))
		})
	}
}

func TestOrderByFilter_Parse(t *testing.T) {
	tests := []struct {
		name              string
		raw               string
		expectedClauses   []OrderByClause
		expectedIsEmpty   bool
	}{
		{
			name:            "empty string",
			raw:             "",
			expectedClauses: []OrderByClause{},
			expectedIsEmpty: true,
		},
		{
			name: "single field default direction",
			raw:  "name",
			expectedClauses: []OrderByClause{
				{Column: "name", Direction: SortAsc},
			},
			expectedIsEmpty: false,
		},
		{
			name: "single field with asc",
			raw:  "name asc",
			expectedClauses: []OrderByClause{
				{Column: "name", Direction: SortAsc},
			},
			expectedIsEmpty: false,
		},
		{
			name: "single field with desc",
			raw:  "name desc",
			expectedClauses: []OrderByClause{
				{Column: "name", Direction: SortDesc},
			},
			expectedIsEmpty: false,
		},
		{
			name: "single field with colon asc",
			raw:  "name:asc",
			expectedClauses: []OrderByClause{
				{Column: "name", Direction: SortAsc},
			},
			expectedIsEmpty: false,
		},
		{
			name: "single field with colon desc",
			raw:  "created_at:desc",
			expectedClauses: []OrderByClause{
				{Column: "created_at", Direction: SortDesc},
			},
			expectedIsEmpty: false,
		},
		{
			name: "single field with minus prefix",
			raw:  "-created_at",
			expectedClauses: []OrderByClause{
				{Column: "created_at", Direction: SortDesc},
			},
			expectedIsEmpty: false,
		},
		{
			name: "multiple fields with spaces",
			raw:  "name asc, created_at desc",
			expectedClauses: []OrderByClause{
				{Column: "name", Direction: SortAsc},
				{Column: "created_at", Direction: SortDesc},
			},
			expectedIsEmpty: false,
		},
		{
			name: "multiple fields with colons",
			raw:  "name:asc,age:desc,created_at:asc",
			expectedClauses: []OrderByClause{
				{Column: "name", Direction: SortAsc},
				{Column: "age", Direction: SortDesc},
				{Column: "created_at", Direction: SortAsc},
			},
			expectedIsEmpty: false,
		},
		{
			name: "mixed delimiters",
			raw:  "name asc, age:desc, -created_at",
			expectedClauses: []OrderByClause{
				{Column: "name", Direction: SortAsc},
				{Column: "age", Direction: SortDesc},
				{Column: "created_at", Direction: SortDesc},
			},
			expectedIsEmpty: false,
		},
		{
			name: "case insensitive directions",
			raw:  "name ASC, age DESC",
			expectedClauses: []OrderByClause{
				{Column: "name", Direction: SortAsc},
				{Column: "age", Direction: SortDesc},
			},
			expectedIsEmpty: false,
		},
		{
			name: "whitespace handling",
			raw:  "  name  asc  ,  created_at  desc  ",
			expectedClauses: []OrderByClause{
				{Column: "name", Direction: SortAsc},
				{Column: "created_at", Direction: SortDesc},
			},
			expectedIsEmpty: false,
		},
		{
			name: "invalid direction defaults to asc",
			raw:  "name invalid",
			expectedClauses: []OrderByClause{
				{Column: "name", Direction: SortAsc},
			},
			expectedIsEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ob := NewOrderByFilter("")
			ob.Parse(tt.raw)

			clauses := ob.GetClauses()
			assert.Equal(t, len(tt.expectedClauses), len(clauses))
			assert.Equal(t, tt.expectedIsEmpty, ob.IsEmpty())

			for i, expectedClause := range tt.expectedClauses {
				if i < len(clauses) {
					assert.Equal(t, expectedClause.Column, clauses[i].Column)
					assert.Equal(t, expectedClause.Direction, clauses[i].Direction)
				}
			}
		})
	}
}

func TestOrderByFilter_ExtractOrderByFromQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "empty query",
			query:    "",
			expected: "",
		},
		{
			name:     "query with order_by parameter",
			query:    "order_by=name asc",
			expected: "name asc",
		},
		{
			name:     "query with leading question mark",
			query:    "?order_by=name desc&page=1",
			expected: "name desc",
		},
		{
			name:     "query without leading question mark",
			query:    "filter=active&order_by=created_at asc&page_size=10",
			expected: "created_at asc",
		},
		{
			name:     "query with orderby parameter (no underscore)",
			query:    "orderby=name asc",
			expected: "name asc",
		},
		{
			name:     "query with sort parameter",
			query:    "sort=name desc",
			expected: "name desc",
		},
		{
			name:     "case insensitive parameter names",
			query:    "ORDER_BY=name asc",
			expected: "name asc",
		},
		{
			name:     "complex query with multiple parameters",
			query:    "filter=status=active&page=1&order_by=name asc,created_at desc&page_size=20",
			expected: "name asc,created_at desc",
		},
		{
			name:     "query with no order_by parameter",
			query:    "filter=active&page=1&page_size=10",
			expected: "",
		},
		{
			name:     "query with empty order_by parameter",
			query:    "order_by=&page=1",
			expected: "",
		},
		{
			name:     "query with malformed parameters",
			query:    "filter&order_by=name asc&page",
			expected: "name asc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ob := &OrderByFilter{}
			result := ob.extractOrderByFromQuery(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrderByFilter_Apply(t *testing.T) {
	db := setupOrderByTestDB(t)

	tests := []struct {
		name           string
		raw            string
		allowedColumns []string
		expectedIDs    []uint
		description    string
	}{
		{
			name:        "empty order by",
			raw:         "",
			expectedIDs: []uint{1, 2, 3, 4}, // default order (insertion order)
			description: "should return records in insertion order",
		},
		{
			name:        "order by name asc",
			raw:         "name asc",
			expectedIDs: []uint{2, 3, 1, 4}, // alice, bob, charlie, diana
			description: "should order by name ascending",
		},
		{
			name:        "order by name desc",
			raw:         "name desc",
			expectedIDs: []uint{4, 1, 3, 2}, // diana, charlie, bob, alice
			description: "should order by name descending",
		},
		{
			name:        "order by age asc",
			raw:         "age asc",
			expectedIDs: []uint{2, 4, 3, 1}, // alice(25), diana(28), bob(30), charlie(35)
			description: "should order by age ascending",
		},
		{
			name:        "order by age desc",
			raw:         "age desc",
			expectedIDs: []uint{1, 3, 4, 2}, // charlie(35), bob(30), diana(28), alice(25)
			description: "should order by age descending",
		},
		{
			name:        "order by created_at asc",
			raw:         "created_at asc",
			expectedIDs: []uint{2, 3, 1, 4}, // 1000, 2000, 3000, 4000
			description: "should order by created_at ascending",
		},
		{
			name:        "multiple order fields",
			raw:         "age asc, name desc",
			expectedIDs: []uint{2, 4, 3, 1}, // age asc: 25, 28, 30, 35 - within same age, name desc
			description: "should order by age asc, then name desc",
		},
		{
			name:        "order with colon delimiter",
			raw:         "name:desc",
			expectedIDs: []uint{4, 1, 3, 2}, // diana, charlie, bob, alice
			description: "should parse colon delimiter",
		},
		{
			name:        "order with minus prefix",
			raw:         "-age",
			expectedIDs: []uint{1, 3, 4, 2}, // charlie(35), bob(30), diana(28), alice(25)
			description: "should parse minus prefix as desc",
		},
		{
			name:           "order with allowed columns",
			raw:            "name asc, age desc",
			allowedColumns: []string{"name"}, // only allow name column
			expectedIDs:    []uint{2, 3, 1, 4}, // only name ordering applied
			description:    "should only apply allowed columns",
		},
		{
			name:           "order with no allowed columns",
			raw:            "name asc",
			allowedColumns: []string{"other_column"}, // disallow name column
			expectedIDs:    []uint{1, 2, 3, 4}, // no ordering applied
			description:    "should ignore disallowed columns",
		},
		{
			name:        "query string format",
			raw:         "?order_by=age desc",
			expectedIDs: []uint{1, 3, 4, 2}, // charlie(35), bob(30), diana(28), alice(25)
			description: "should parse query string format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ob := NewOrderByFilter(tt.raw)

			var results []TestOrderByModel
			query := ob.Apply(db.Model(&TestOrderByModel{}), tt.allowedColumns...)
			err := query.Find(&results).Error
			require.NoError(t, err)

			assert.Equal(t, len(tt.expectedIDs), len(results), tt.description)

			actualIDs := make([]uint, len(results))
			for i, result := range results {
				actualIDs[i] = result.ID
			}
			assert.Equal(t, tt.expectedIDs, actualIDs, tt.description)
		})
	}
}

func TestOrderByFilter_ApplyWithNilDB(t *testing.T) {
	ob := NewOrderByFilter("name asc")

	result := ob.Apply(nil)
	assert.Nil(t, result)
}

func TestOrderByFilter_ApplyWithEmptyFilter(t *testing.T) {
	db := setupOrderByTestDB(t)
	ob := NewOrderByFilter("")

	var results []TestOrderByModel
	query := ob.Apply(db.Model(&TestOrderByModel{}))
	err := query.Find(&results).Error
	require.NoError(t, err)

	// Should return results in default order (no ORDER BY clause added)
	assert.Equal(t, 4, len(results))
}

func TestOrderByFilter_GetClauses(t *testing.T) {
	ob := NewOrderByFilter("name asc, age desc")
	clauses := ob.GetClauses()

	assert.Equal(t, 2, len(clauses))
	assert.Equal(t, "name", clauses[0].Column)
	assert.Equal(t, SortAsc, clauses[0].Direction)
	assert.Equal(t, "age", clauses[1].Column)
	assert.Equal(t, SortDesc, clauses[1].Direction)
}

func TestOrderByFilter_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected bool
	}{
		{
			name:     "empty string",
			raw:      "",
			expected: true,
		},
		{
			name:     "whitespace only",
			raw:      "   ",
			expected: true,
		},
		{
			name:     "valid order clause",
			raw:      "name asc",
			expected: false,
		},
		{
			name:     "query string with no order_by",
			raw:      "?filter=active&page=1",
			expected: true,
		},
		{
			name:     "query string with order_by",
			raw:      "?order_by=name asc",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ob := NewOrderByFilter(tt.raw)
			assert.Equal(t, tt.expected, ob.IsEmpty())
		})
	}
}

func TestSortDirection_Constants(t *testing.T) {
	assert.Equal(t, SortDirection("asc"), SortAsc)
	assert.Equal(t, SortDirection("desc"), SortDesc)
}

func TestOrderByClause_Struct(t *testing.T) {
	clause := OrderByClause{
		Column:    "test_column",
		Direction: SortDesc,
	}

	assert.Equal(t, "test_column", clause.Column)
	assert.Equal(t, SortDesc, clause.Direction)
}

func TestOrderByFilter_ComplexScenarios(t *testing.T) {
	db := setupOrderByTestDB(t)

	tests := []struct {
		name        string
		raw         string
		expectedIDs []uint
		description string
	}{
		{
			name:        "multiple mixed delimiters",
			raw:         "name asc, age:desc, -created_at",
			expectedIDs: []uint{2, 3, 1, 4}, // complex multi-field sorting
			description: "should handle mixed delimiter formats",
		},
		{
			name:        "case insensitive directions",
			raw:         "name ASC, age DESC",
			expectedIDs: []uint{2, 3, 1, 4}, // alice(25), bob(30), charlie(35), diana(28) - name asc, then age desc
			description: "should handle case insensitive direction keywords",
		},
		{
			name:        "extra whitespace",
			raw:         "  name   asc  ,   age   desc  ",
			expectedIDs: []uint{2, 3, 1, 4}, // same as above
			description: "should handle extra whitespace gracefully",
		},
		{
			name:        "single field multiple spaces",
			raw:         "name    asc",
			expectedIDs: []uint{2, 3, 1, 4},
			description: "should handle multiple spaces in field definition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ob := NewOrderByFilter(tt.raw)

			var results []TestOrderByModel
			query := ob.Apply(db.Model(&TestOrderByModel{}))
			err := query.Find(&results).Error
			require.NoError(t, err)

			actualIDs := make([]uint, len(results))
			for i, result := range results {
				actualIDs[i] = result.ID
			}
			assert.Equal(t, tt.expectedIDs, actualIDs, tt.description)
		})
	}
}

func TestOrderByFilter_EdgeCases(t *testing.T) {
	tests := []struct {
		name            string
		raw             string
		expectedClauses int
		description     string
	}{
		{
			name:            "empty field names",
			raw:             ", , ,",
			expectedClauses: 0,
			description:     "should ignore empty field names",
		},
		{
			name:            "fields with only minus",
			raw:             "-",
			expectedClauses: 0,
			description:     "should ignore fields that are only minus",
		},
		{
			name:            "fields with only colon",
			raw:             ":",
			expectedClauses: 0,
			description:     "should ignore fields that are only colon",
		},
		{
			name:            "mixed valid and invalid",
			raw:             "name asc, , age desc, -",
			expectedClauses: 2,
			description:     "should parse valid fields and ignore invalid ones",
		},
		{
			name:            "colon without direction",
			raw:             "name:",
			expectedClauses: 1,
			description:     "should default to asc when direction is missing after colon",
		},
		{
			name:            "multiple colons",
			raw:             "name:asc:extra",
			expectedClauses: 1,
			description:     "should handle multiple colons gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ob := NewOrderByFilter(tt.raw)
			assert.Equal(t, tt.expectedClauses, len(ob.GetClauses()), tt.description)
		})
	}
}