package restapi

type TokenRoleClaimValidation struct {
	Name   string
	exists bool
}

func (s *TokenRoleClaimValidation) Exists() bool {
	return s.exists
}

func (s *TokenRoleClaimValidation) SetExists(exists bool) {
	s.exists = exists
}

type TokenRoleClaimValidationList []*TokenRoleClaimValidation

func (s TokenRoleClaimValidationList) Exists() bool {
	for _, item := range s {
		if !item.Exists() {
			return false
		}
	}
	return true
}

func (s TokenRoleClaimValidationList) ExistsAny() bool {
	for _, item := range s {
		if item.Exists() {
			return true
		}
	}

	return false
}

func (s TokenRoleClaimValidationList) Evaluate(operation ComparisonOperation) bool {
	op := normalizeComparisonOperation(operation)
	if op == ComparisonOperationOr {
		return s.ExistsAny()
	}

	return s.Exists()
}

func (s TokenRoleClaimValidationList) GetFailed() []string {
	failed := []string{}
	for _, item := range s {
		if !item.Exists() {
			failed = append(failed, item.Name)
		}
	}

	return failed
}
