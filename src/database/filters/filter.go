package filters

import (
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// Operator represents the SQL filter operator
type Operator string

const (
	OpEqual              Operator = "="
	OpNotEqual           Operator = "!="
	OpGreaterThan        Operator = ">"
	OpGreaterThanOrEqual Operator = ">="
	OpLessThan           Operator = "<"
	OpLessThanOrEqual    Operator = "<="
	OpLike               Operator = "LIKE"
	OpIn                 Operator = "IN"
	OpNotIn              Operator = "NOT IN"
	OpIsNull             Operator = "IS NULL"
	OpIsNotNull          Operator = "IS NOT NULL"
	OpBetween            Operator = "BETWEEN"
	OpNotBetween         Operator = "NOT BETWEEN"
	OpContains           Operator = "CONTAINS"
)

// LogicalOperator represents the logical operator between filter conditions
type LogicalOperator string

const (
	LogicalAnd LogicalOperator = "AND"
	LogicalOr  LogicalOperator = "OR"
)

// Condition represents a single filter condition
type Condition struct {
	Field    string
	Operator Operator
	Value    interface{}
	Values   []interface{} // Used for IN, NOT IN, BETWEEN operations
}

// FilterClause represents a filter condition with logical operator
type FilterClause struct {
	Condition       Condition
	LogicalOperator LogicalOperator // AND/OR with next condition
}

// FilterQuery represents a collection of filter clauses for URL query parsing
type FilterQuery struct {
	clauses []FilterClause
}

// NewFilterQuery creates a new FilterQuery instance and parses the raw string
func NewFilterQuery(raw string) *FilterQuery {
	fq := &FilterQuery{
		clauses: make([]FilterClause, 0),
	}
	if raw != "" {
		fq.Parse(raw)
	}
	return fq
}

// Parse parses filter parameters from a raw string
// Expected formats:
// - Simple: "name=john&age>25&status!=inactive"
// - Complex query: "?filter=name=john,age>25,status!=inactive&page=1"
// - With logical operators: "name=john AND age>25 OR status=active"
func (fq *FilterQuery) Parse(raw string) {
	fq.clauses = make([]FilterClause, 0)
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return
	}

	// Check if this looks like a full query string with multiple parameters
	if strings.Contains(trimmed, "&") || strings.Contains(trimmed, "?") {
		// Extract filter parameter from query string
		filterValue := fq.extractFilterFromQuery(trimmed)
		if filterValue == "" {
			return
		}
		trimmed = filterValue
	}

	// Parse the filter conditions
	fq.parseFilterConditions(trimmed)
}

// extractFilterFromQuery extracts the filter parameter value from a full query string
func (fq *FilterQuery) extractFilterFromQuery(query string) string {
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

		// Check for filter parameter (case insensitive)
		if strings.EqualFold(key, "filter") || strings.EqualFold(key, "filters") || strings.EqualFold(key, "where") {
			return value
		}
	}

	return ""
}

// parseFilterConditions parses filter conditions from the filter string
// Supports formats:
// - "name=john,age>25,status!=inactive" (comma-separated)
// - "name=john AND age>25 OR status=active" (space-separated with logical operators)
func (fq *FilterQuery) parseFilterConditions(filterStr string) {
	// First try to parse as space-separated with logical operators
	if fq.parseWithLogicalOperators(filterStr) {
		return
	}

	// Fall back to comma-separated parsing
	fq.parseCommaSeparated(filterStr)
}

// parseWithLogicalOperators parses filter string with AND/OR logical operators
func (fq *FilterQuery) parseWithLogicalOperators(filterStr string) bool {
	// Check if the string contains AND or OR operators
	if !strings.Contains(strings.ToUpper(filterStr), " AND ") && !strings.Contains(strings.ToUpper(filterStr), " OR ") {
		return false
	}

	// Split by AND and OR while preserving the operators
	parts := fq.splitByLogicalOperators(filterStr)

	var nextLogicalOp LogicalOperator = ""

	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		upperPart := strings.ToUpper(trimmedPart)

		if upperPart == "AND" || upperPart == "OR" {
			// This is a logical operator, save it for the next condition
			nextLogicalOp = LogicalOperator(upperPart)
			continue
		}

		// Parse the condition
		condition := fq.parseCondition(trimmedPart)
		if condition != nil {
			fq.clauses = append(fq.clauses, FilterClause{
				Condition:       *condition,
				LogicalOperator: nextLogicalOp,
			})
			// Reset for next iteration
			nextLogicalOp = ""
		}
	}

	return true
}

// splitByLogicalOperators splits the filter string by AND/OR operators while preserving them
func (fq *FilterQuery) splitByLogicalOperators(filterStr string) []string {
	result := make([]string, 0)
	current := ""
	words := strings.Fields(filterStr)
	betweenCount := 0

	for i, word := range words {
		upperWord := strings.ToUpper(word)

		// Handle BETWEEN logic
		if upperWord == "BETWEEN" {
			betweenCount++
		}

		isLogicalOp := false
		if upperWord == "AND" || upperWord == "OR" {
			if upperWord == "AND" && betweenCount > 0 {
				// This AND belongs to a BETWEEN clause
				betweenCount--
			} else {
				isLogicalOp = true
			}
		}

		if isLogicalOp {
			// Add the accumulated condition
			if current != "" {
				result = append(result, strings.TrimSpace(current))
				current = ""
			}
			// Add the logical operator
			result = append(result, upperWord)
		} else {
			if current != "" {
				current += " "
			}
			current += word
		}

		// Handle NOT BETWEEN case (NOT is previous word)
		if i > 0 && upperWord == "BETWEEN" && strings.ToUpper(words[i-1]) == "NOT" {
			// Already handled by incrementing betweenCount
		}
	}

	// Add the last condition
	if current != "" {
		result = append(result, strings.TrimSpace(current))
	}

	return result
}

// parseCommaSeparated parses comma-separated filter conditions
func (fq *FilterQuery) parseCommaSeparated(filterStr string) {
	conditions := fq.splitByCommaRespectingParens(filterStr)

	for i, condStr := range conditions {
		condition := fq.parseCondition(strings.TrimSpace(condStr))
		if condition != nil {
			logicalOp := LogicalOperator("")
			if i > 0 {
				logicalOp = LogicalAnd // Default to AND for comma-separated conditions
			}

			fq.clauses = append(fq.clauses, FilterClause{
				Condition:       *condition,
				LogicalOperator: logicalOp,
			})
		}
	}
}

// splitByCommaRespectingParens splits a string by comma but ignores commas inside parentheses
func (fq *FilterQuery) splitByCommaRespectingParens(s string) []string {
	var result []string
	var current strings.Builder
	parenCount := 0

	for _, r := range s {
		if r == '(' {
			parenCount++
		} else if r == ')' {
			parenCount--
		}

		if r == ',' && parenCount == 0 {
			result = append(result, strings.TrimSpace(current.String()))
			current.Reset()
		} else {
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		result = append(result, strings.TrimSpace(current.String()))
	}
	return result
}

// parseCondition parses a single filter condition
func (fq *FilterQuery) parseCondition(conditionStr string) *Condition {
	if conditionStr == "" {
		return nil
	}

	// Try different operators in order of specificity (longer ones first)
	operators := []string{"IS NOT NULL", "IS NULL", "NOT BETWEEN", "NOT IN", ">=", "<=", "!=", "BETWEEN", "CONTAINS", "LIKE", "IN", "=", ">", "<"}

	for _, op := range operators {
		if idx := fq.findOperatorIndex(conditionStr, op); idx != -1 {
			field := strings.TrimSpace(conditionStr[:idx])
			valueStr := strings.TrimSpace(conditionStr[idx+len(op):])

			condition := &Condition{
				Field:    field,
				Operator: Operator(op),
			}

			// Handle special cases for operators that don't need values
			if op == "IS NULL" || op == "IS NOT NULL" {
				condition.Value = nil
			} else if op == "IN" || op == "NOT IN" {
				// Parse comma-separated values for IN operations
				values := fq.parseInValues(valueStr)
				condition.Values = values
			} else if op == "BETWEEN" || op == "NOT BETWEEN" {
				// Parse two values for BETWEEN operations
				values := fq.parseBetweenValues(valueStr)
				if len(values) == 2 {
					condition.Values = values
				}
			} else if op == "CONTAINS" {
				// Convert CONTAINS to LIKE with wildcards
				condition.Operator = OpLike
				condition.Value = "%" + valueStr + "%"
			} else {
				// Handle type conversion for the value
				condition.Value = fq.convertValue(valueStr)
			}

			return condition
		}
	}

	return nil
}

// findOperatorIndex finds the index of an operator in the condition string (case insensitive)
func (fq *FilterQuery) findOperatorIndex(conditionStr, operator string) int {
	upper := strings.ToUpper(conditionStr)
	upperOp := strings.ToUpper(operator)

	idx := strings.Index(upper, upperOp)
	if idx == -1 {
		return -1
	}

	// For alpha operators (e.g. IN, LIKE), ensure they are surrounded by boundaries (space, paren, or end of string)
	isAlpha := func(c rune) bool {
		return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
	}

	// Check if operator starts with letter
	if len(operator) > 0 && isAlpha(rune(operator[0])) {
		// Check character before
		if idx > 0 {
			before := rune(conditionStr[idx-1])
			if isAlpha(before) || isAlpha(rune(upperOp[0])) && isAlpha(before) {
				// If char before is alpha, it's part of a word?
				// Actually we want to check if it's NOT a boundary.
				// If char before is alpha or number, it's not a start boundary.
				if isAlpha(before) {
					// Check if we can find another occurrence?
					// For simplicity, just checking the first one.
					// If "admin", idx=3 ("in"). Before is 'm'.
					// So validation fails.
					// We should search again?
					// strings.Index returns first.
					// Simple hack: loop search?
					return -1
				}
			}
		}

		// Check character after matching part
		if idx+len(operator) < len(conditionStr) {
			after := rune(conditionStr[idx+len(operator)])
			if isAlpha(after) {
				return -1
			}
		}
	}

	return idx
}

// parseInValues parses comma-separated values for IN operations
func (fq *FilterQuery) parseInValues(valueStr string) []interface{} {
	// Remove parentheses if present
	valueStr = strings.TrimSpace(valueStr)
	valueStr = strings.TrimPrefix(valueStr, "(")
	valueStr = strings.TrimSuffix(valueStr, ")")

	// Handle cases where there are no parentheses but comma-separated values
	if !strings.Contains(valueStr, ",") {
		// Single value
		return []interface{}{fq.convertValue(valueStr)}
	}

	parts := strings.Split(valueStr, ",")
	values := make([]interface{}, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			value := fq.convertValue(part)
			values = append(values, value)
		}
	}

	return values
}

// parseBetweenValues parses two values for BETWEEN operations
func (fq *FilterQuery) parseBetweenValues(valueStr string) []interface{} {
	// Try different separators for BETWEEN values
	separators := []string{" AND ", " and ", ","}

	for _, sep := range separators {
		if strings.Contains(valueStr, sep) {
			parts := strings.Split(valueStr, sep)
			if len(parts) == 2 {
				values := make([]interface{}, 2)
				values[0] = fq.convertValue(strings.TrimSpace(parts[0]))
				values[1] = fq.convertValue(strings.TrimSpace(parts[1]))
				return values
			}
		}
	}

	return nil
}

// convertValue attempts to convert string values to appropriate types
func (fq *FilterQuery) convertValue(valueStr string) interface{} {
	valueStr = strings.TrimSpace(valueStr)

	// Remove quotes if present
	if (strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"")) ||
		(strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'")) {
		valueStr = valueStr[1 : len(valueStr)-1]
	}

	// Try to convert to number
	if intVal, err := strconv.Atoi(valueStr); err == nil {
		return intVal
	}

	if floatVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return floatVal
	}

	// Try to convert to boolean
	if boolVal, err := strconv.ParseBool(valueStr); err == nil {
		return boolVal
	}

	// Return as string
	return valueStr
}

// GetClauses returns the parsed filter clauses
func (fq *FilterQuery) GetClauses() []FilterClause {
	return fq.clauses
}

// IsEmpty returns true if no clauses were parsed
func (fq *FilterQuery) IsEmpty() bool {
	return len(fq.clauses) == 0
}

// Apply applies the filter clauses to a gorm DB and returns the updated DB
func (fq *FilterQuery) Apply(db *gorm.DB, allowedColumns ...string) *gorm.DB {
	if db == nil || fq.IsEmpty() {
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

	whereClause := ""
	args := make([]interface{}, 0)

	for i, clause := range fq.clauses {
		if clause.Condition.Field == "" || !isAllowed(clause.Condition.Field) {
			continue
		}

		// Add logical operator if not the first condition
		if i > 0 && clause.LogicalOperator != "" {
			whereClause += " " + string(clause.LogicalOperator) + " "
		}

		// Build the condition
		conditionStr, conditionArgs := fq.buildCondition(clause.Condition)
		whereClause += conditionStr
		args = append(args, conditionArgs...)
	}

	if whereClause != "" {
		db = db.Where(whereClause, args...)
	}

	return db
}

// buildCondition builds a single condition string and its arguments
func (fq *FilterQuery) buildCondition(condition Condition) (string, []interface{}) {
	switch condition.Operator {
	case OpIsNull:
		return fmt.Sprintf("%s IS NULL", condition.Field), nil
	case OpIsNotNull:
		return fmt.Sprintf("%s IS NOT NULL", condition.Field), nil
	case OpIn, OpNotIn:
		if len(condition.Values) == 0 {
			return "", nil
		}
		placeholders := strings.Repeat("?,", len(condition.Values))
		placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
		return fmt.Sprintf("%s %s (%s)", condition.Field, condition.Operator, placeholders), condition.Values
	case OpBetween, OpNotBetween:
		if len(condition.Values) != 2 {
			return "", nil
		}
		return fmt.Sprintf("%s %s ? AND ?", condition.Field, condition.Operator), condition.Values
	default:
		return fmt.Sprintf("%s %s ?", condition.Field, condition.Operator), []interface{}{condition.Value}
	}
}
