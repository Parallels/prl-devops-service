package filters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewFilterQuery(t *testing.T) {
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
			name:     "simple filter",
			raw:      "name=john",
			expected: 1,
		},
		{
			name:     "multiple filters comma-separated",
			raw:      "name=john,age>25",
			expected: 2,
		},
		{
			name:     "filter with query string",
			raw:      "?filter=name=john&page=1",
			expected: 1,
		},
		{
			name:     "complex query with filters",
			raw:      "name=john AND age>25",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFilterQuery(tt.raw)
			assert.Equal(t, tt.expected, len(result.GetClauses()))
		})
	}
}

func TestFilterQuery_Parse(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected []FilterClause
	}{
		{
			name: "empty string",
			raw:  "",
			expected: []FilterClause{},
		},
		{
			name: "simple equals",
			raw:  "name=john",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "name",
						Operator: OpEqual,
						Value:    "john",
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "not equals",
			raw:  "status!=inactive",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "status",
						Operator: OpNotEqual,
						Value:    "inactive",
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "greater than with number",
			raw:  "age>25",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "age",
						Operator: OpGreaterThan,
						Value:    25,
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "greater than or equal",
			raw:  "score>=80",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "score",
						Operator: OpGreaterThanOrEqual,
						Value:    80,
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "less than",
			raw:  "price<100.50",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "price",
						Operator: OpLessThan,
						Value:    100.50,
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "less than or equal",
			raw:  "discount<=20",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "discount",
						Operator: OpLessThanOrEqual,
						Value:    20,
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "LIKE operator",
			raw:  "title LIKE %search%",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "title",
						Operator: OpLike,
						Value:    "%search%",
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "CONTAINS operator",
			raw:  "description CONTAINS keyword",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "description",
						Operator: OpLike,
						Value:    "%keyword%",
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "IN operator",
			raw:  "category IN (tech,science,art)",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "category",
						Operator: OpIn,
						Values:   []interface{}{"tech", "science", "art"},
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "NOT IN operator",
			raw:  "status NOT IN (deleted,archived)",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "status",
						Operator: OpNotIn,
						Values:   []interface{}{"deleted", "archived"},
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "IS NULL operator",
			raw:  "deleted_at IS NULL",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "deleted_at",
						Operator: OpIsNull,
						Value:    nil,
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "IS NOT NULL operator",
			raw:  "email IS NOT NULL",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "email",
						Operator: OpIsNotNull,
						Value:    nil,
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "BETWEEN operator",
			raw:  "age BETWEEN 18 AND 65",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "age",
						Operator: OpBetween,
						Values:   []interface{}{18, 65},
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "NOT BETWEEN operator",
			raw:  "score NOT BETWEEN 0 AND 50",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "score",
						Operator: OpNotBetween,
						Values:   []interface{}{0, 50},
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "comma-separated multiple conditions",
			raw:  "name=john,age>25,status=active",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "name",
						Operator: OpEqual,
						Value:    "john",
					},
					LogicalOperator: "",
				},
				{
					Condition: Condition{
						Field:    "age",
						Operator: OpGreaterThan,
						Value:    25,
					},
					LogicalOperator: LogicalAnd,
				},
				{
					Condition: Condition{
						Field:    "status",
						Operator: OpEqual,
						Value:    "active",
					},
					LogicalOperator: LogicalAnd,
				},
			},
		},
		{
			name: "AND logical operator",
			raw:  "name=john AND age>25",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "name",
						Operator: OpEqual,
						Value:    "john",
					},
					LogicalOperator: "",
				},
				{
					Condition: Condition{
						Field:    "age",
						Operator: OpGreaterThan,
						Value:    25,
					},
					LogicalOperator: LogicalAnd,
				},
			},
		},
		{
			name: "OR logical operator",
			raw:  "status=active OR status=pending",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "status",
						Operator: OpEqual,
						Value:    "active",
					},
					LogicalOperator: "",
				},
				{
					Condition: Condition{
						Field:    "status",
						Operator: OpEqual,
						Value:    "pending",
					},
					LogicalOperator: LogicalOr,
				},
			},
		},
		{
			name: "mixed AND OR operators",
			raw:  "name=john AND age>25 OR status=admin",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "name",
						Operator: OpEqual,
						Value:    "john",
					},
					LogicalOperator: "",
				},
				{
					Condition: Condition{
						Field:    "age",
						Operator: OpGreaterThan,
						Value:    25,
					},
					LogicalOperator: LogicalAnd,
				},
				{
					Condition: Condition{
						Field:    "status",
						Operator: OpEqual,
						Value:    "admin",
					},
					LogicalOperator: LogicalOr,
				},
			},
		},
		{
			name: "boolean value",
			raw:  "active=true",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "active",
						Operator: OpEqual,
						Value:    true,
					},
					LogicalOperator: "",
				},
			},
		},
		{
			name: "quoted string value",
			raw:  "name=\"John Doe\"",
			expected: []FilterClause{
				{
					Condition: Condition{
						Field:    "name",
						Operator: OpEqual,
						Value:    "John Doe",
					},
					LogicalOperator: "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &FilterQuery{}
			fq.Parse(tt.raw)
			result := fq.GetClauses()
			
			assert.Equal(t, len(tt.expected), len(result), "Number of clauses should match")
			
			for i, expectedClause := range tt.expected {
				if i < len(result) {
					assert.Equal(t, expectedClause.Condition.Field, result[i].Condition.Field, "Field should match")
					assert.Equal(t, expectedClause.Condition.Operator, result[i].Condition.Operator, "Operator should match")
					assert.Equal(t, expectedClause.Condition.Value, result[i].Condition.Value, "Value should match")
					assert.Equal(t, expectedClause.Condition.Values, result[i].Condition.Values, "Values should match")
					assert.Equal(t, expectedClause.LogicalOperator, result[i].LogicalOperator, "Logical operator should match")
				}
			}
		})
	}
}

func TestFilterQuery_ExtractFilterFromQuery(t *testing.T) {
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
			name:     "query with filter parameter",
			query:    "?filter=name=john&page=1",
			expected: "name=john",
		},
		{
			name:     "query without filter parameter",
			query:    "?page=1&limit=10",
			expected: "",
		},
		{
			name:     "query with filters parameter",
			query:    "filters=status=active,age>25&sort=name",
			expected: "status=active,age>25",
		},
		{
			name:     "query with where parameter",
			query:    "where=deleted_at IS NULL&order=created_at",
			expected: "deleted_at IS NULL",
		},
		{
			name:     "complex filter in query",
			query:    "?filter=name=john AND age>25&page=2&limit=20",
			expected: "name=john AND age>25",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &FilterQuery{}
			result := fq.extractFilterFromQuery(tt.query)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterQuery_ConvertValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "integer",
			input:    "123",
			expected: 123,
		},
		{
			name:     "float",
			input:    "123.45",
			expected: 123.45,
		},
		{
			name:     "boolean true",
			input:    "true",
			expected: true,
		},
		{
			name:     "boolean false",
			input:    "false",
			expected: false,
		},
		{
			name:     "quoted string",
			input:    "\"hello world\"",
			expected: "hello world",
		},
		{
			name:     "single quoted string",
			input:    "'hello world'",
			expected: "hello world",
		},
		{
			name:     "regular string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string with spaces",
			input:    "  hello world  ",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &FilterQuery{}
			result := fq.convertValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterQuery_GetClauses(t *testing.T) {
	fq := NewFilterQuery("name=john,age>25")
	clauses := fq.GetClauses()
	
	assert.Equal(t, 2, len(clauses))
	assert.Equal(t, "name", clauses[0].Condition.Field)
	assert.Equal(t, "age", clauses[1].Condition.Field)
}

func TestFilterQuery_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		fq       *FilterQuery
		expected bool
	}{
		{
			name:     "empty filter",
			fq:       NewFilterQuery(""),
			expected: true,
		},
		{
			name:     "filter with clauses",
			fq:       NewFilterQuery("name=john"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fq.IsEmpty()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterQuery_Apply(t *testing.T) {
	// Setup in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	tests := []struct {
		name           string
		fq             *FilterQuery
		db             *gorm.DB
		allowedColumns []string
		shouldChange   bool
	}{
		{
			name:         "nil database",
			fq:           NewFilterQuery("name=john"),
			db:           nil,
			shouldChange: false,
		},
		{
			name:         "empty filter",
			fq:           NewFilterQuery(""),
			db:           db,
			shouldChange: false,
		},
		{
			name:         "valid filter",
			fq:           NewFilterQuery("name=john"),
			db:           db,
			shouldChange: true,
		},
		{
			name:           "filter with allowed columns",
			fq:             NewFilterQuery("name=john,email=test@example.com"),
			db:             db,
			allowedColumns: []string{"name"},
			shouldChange:   true,
		},
		{
			name:           "filter with disallowed columns",
			fq:             NewFilterQuery("password=secret"),
			db:             db,
			allowedColumns: []string{"name", "email"},
			shouldChange:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fq.Apply(tt.db, tt.allowedColumns...)
			
			if tt.name == "nil database" {
				assert.Nil(t, result)
			} else if !tt.shouldChange {
				assert.Equal(t, tt.db, result)
			} else {
				assert.NotNil(t, result)
				// The result should be a different DB instance with applied filters
				assert.NotEqual(t, tt.db, result)
			}
		})
	}
}

func TestFilterQuery_BuildCondition(t *testing.T) {
	tests := []struct {
		name              string
		condition         Condition
		expectedQuery     string
		expectedArgsCount int
	}{
		{
			name: "equals condition",
			condition: Condition{
				Field:    "name",
				Operator: OpEqual,
				Value:    "john",
			},
			expectedQuery:     "name = ?",
			expectedArgsCount: 1,
		},
		{
			name: "IS NULL condition",
			condition: Condition{
				Field:    "deleted_at",
				Operator: OpIsNull,
			},
			expectedQuery:     "deleted_at IS NULL",
			expectedArgsCount: 0,
		},
		{
			name: "IS NOT NULL condition",
			condition: Condition{
				Field:    "email",
				Operator: OpIsNotNull,
			},
			expectedQuery:     "email IS NOT NULL",
			expectedArgsCount: 0,
		},
		{
			name: "IN condition",
			condition: Condition{
				Field:    "status",
				Operator: OpIn,
				Values:   []interface{}{"active", "pending"},
			},
			expectedQuery:     "status IN (?,?)",
			expectedArgsCount: 2,
		},
		{
			name: "NOT IN condition",
			condition: Condition{
				Field:    "status",
				Operator: OpNotIn,
				Values:   []interface{}{"deleted", "archived"},
			},
			expectedQuery:     "status NOT IN (?,?)",
			expectedArgsCount: 2,
		},
		{
			name: "BETWEEN condition",
			condition: Condition{
				Field:    "age",
				Operator: OpBetween,
				Values:   []interface{}{18, 65},
			},
			expectedQuery:     "age BETWEEN ? AND ?",
			expectedArgsCount: 2,
		},
		{
			name: "NOT BETWEEN condition",
			condition: Condition{
				Field:    "score",
				Operator: OpNotBetween,
				Values:   []interface{}{0, 50},
			},
			expectedQuery:     "score NOT BETWEEN ? AND ?",
			expectedArgsCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := &FilterQuery{}
			query, args := fq.buildCondition(tt.condition)
			
			assert.Equal(t, tt.expectedQuery, query)
			assert.Equal(t, tt.expectedArgsCount, len(args))
		})
	}
}

func TestFilterQuery_ComplexScenarios(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(*testing.T, *FilterQuery)
	}{
		{
			name:  "URL query with complex filter",
			input: "?filter=name=john AND age>25 OR status IN (active,pending)&page=1&limit=10",
			validate: func(t *testing.T, fq *FilterQuery) {
				clauses := fq.GetClauses()
				assert.Equal(t, 3, len(clauses))
				
				// First condition: name=john
				assert.Equal(t, "name", clauses[0].Condition.Field)
				assert.Equal(t, OpEqual, clauses[0].Condition.Operator)
				assert.Equal(t, "john", clauses[0].Condition.Value)
				assert.Equal(t, LogicalOperator(""), clauses[0].LogicalOperator)
				
				// Second condition: age>25 AND
				assert.Equal(t, "age", clauses[1].Condition.Field)
				assert.Equal(t, OpGreaterThan, clauses[1].Condition.Operator)
				assert.Equal(t, 25, clauses[1].Condition.Value)
				assert.Equal(t, LogicalAnd, clauses[1].LogicalOperator)
				
				// Third condition: status IN (active,pending) OR
				assert.Equal(t, "status", clauses[2].Condition.Field)
				assert.Equal(t, OpIn, clauses[2].Condition.Operator)
				assert.Equal(t, []interface{}{"active", "pending"}, clauses[2].Condition.Values)
				assert.Equal(t, LogicalOr, clauses[2].LogicalOperator)
			},
		},
		{
			name:  "comma-separated with mixed data types",
			input: "name=john,age>25,active=true,score>=85.5",
			validate: func(t *testing.T, fq *FilterQuery) {
				clauses := fq.GetClauses()
				assert.Equal(t, 4, len(clauses))
				
				assert.Equal(t, "john", clauses[0].Condition.Value)
				assert.Equal(t, 25, clauses[1].Condition.Value)
				assert.Equal(t, true, clauses[2].Condition.Value)
				assert.Equal(t, 85.5, clauses[3].Condition.Value)
			},
		},
		{
			name:  "edge case with empty values",
			input: "name=,status!=",
			validate: func(t *testing.T, fq *FilterQuery) {
				clauses := fq.GetClauses()
				assert.Equal(t, 2, len(clauses))
				
				assert.Equal(t, "", clauses[0].Condition.Value)
				assert.Equal(t, "", clauses[1].Condition.Value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fq := NewFilterQuery(tt.input)
			tt.validate(t, fq)
		})
	}
}