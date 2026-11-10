package data_test

import (
	"testing"

	"github.com/Parallels/prl-devops-service/data"
	"github.com/Parallels/prl-devops-service/data/interfaces"
)

// R-1 guard (#421 D-2): the frozen surface. Any signature drift between the
// 158-method inventory and data.JsonDatabase fails compilation of this file.
var _ interfaces.DatabaseService = (*data.JsonDatabase)(nil)

// D-3 identity proof (#421): startup code compares sentinels by pointer
// identity (startup/seeds/role_claims.go, startup/migrations/add_api_key_own_claims.go),
// so the data aliases must be the identical values relocated to interfaces.
func TestSentinelIdentityPreserved(t *testing.T) {
	pairs := []struct {
		name            string
		dataErr, ifcErr error
	}{
		{"ErrUserNotFound", data.ErrUserNotFound, interfaces.ErrUserNotFound},
		{"ErrRoleAlreadyContainsClaim", data.ErrRoleAlreadyContainsClaim, interfaces.ErrRoleAlreadyContainsClaim},
		{"ErrClaimNotFound", data.ErrClaimNotFound, interfaces.ErrClaimNotFound},
		{"ErrUserAlreadyContainsClaim", data.ErrUserAlreadyContainsClaim, interfaces.ErrUserAlreadyContainsClaim},
		{"ErrDatabaseNotConnected", data.ErrDatabaseNotConnected, interfaces.ErrDatabaseNotConnected},
		{"ErrNotAuthorized", data.ErrNotAuthorized, interfaces.ErrNotAuthorized},
		// stdlib-constructed block (errors.New, no code) included on purpose:
		{"ErrCatalogManagerNotFound", data.ErrCatalogManagerNotFound, interfaces.ErrCatalogManagerNotFound},
	}
	for _, p := range pairs {
		if p.dataErr != p.ifcErr {
			t.Errorf("%s: data alias is not the identical sentinel value", p.name)
		}
	}
}
