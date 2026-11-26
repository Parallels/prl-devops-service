package database

import (
	"testing"

	"github.com/Parallels/prl-devops-service/data/models"
)

// TestAllGormModelsIncludesPlacementDomain guards the boot-registry contract for
// the placement tables, same shape as the settings guard: an unregistered model
// is not a compile error, it is a deployment whose very first create fails with
// "no such table: resource_reservations" — and the store's own tests would only
// report it as twenty opaque automigrate failures.
func TestAllGormModelsIncludesPlacementDomain(t *testing.T) {
	want := map[string]bool{
		"ResourceReservation": false,
		"PlacementGuard":      false,
		"JobLink":             false,
	}

	for _, m := range AllGormModels() {
		switch m.(type) {
		case *models.ResourceReservation:
			want["ResourceReservation"] = true
		case *models.PlacementGuard:
			want["PlacementGuard"] = true
		case *models.JobLink:
			want["JobLink"] = true
		}
	}

	for name, found := range want {
		if !found {
			t.Errorf("AllGormModels missing %s", name)
		}
	}
}
