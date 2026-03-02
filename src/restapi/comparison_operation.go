package restapi

import "strings"

type ComparisonOperation string

const (
	ComparisonOperationAnd ComparisonOperation = "AND"
	ComparisonOperationOr  ComparisonOperation = "OR"
)

func normalizeComparisonOperation(operation ComparisonOperation) ComparisonOperation {
	if strings.EqualFold(string(operation), string(ComparisonOperationOr)) {
		return ComparisonOperationOr
	}

	return ComparisonOperationAnd
}
