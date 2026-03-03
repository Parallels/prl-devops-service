package restapi

import "testing"

func TestTokenRoleClaimValidationList_Evaluate(t *testing.T) {
	validations := TokenRoleClaimValidationList{
		&TokenRoleClaimValidation{Name: "a"},
		&TokenRoleClaimValidation{Name: "b"},
	}

	validations[0].SetExists(true)
	validations[1].SetExists(false)

	if validations.Evaluate(ComparisonOperationAnd) {
		t.Errorf("Expected AND evaluation to fail when one validation is missing")
	}

	if !validations.Evaluate(ComparisonOperationOr) {
		t.Errorf("Expected OR evaluation to pass when one validation exists")
	}
}
