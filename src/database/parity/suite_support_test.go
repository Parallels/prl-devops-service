package parity

// Shared suite support: contexts and canonicalisation helpers.

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
)

func newRoleClaimCtx() basecontext.ApiContext { return basecontext.NewBaseContext() }

// normaliseRoles canonicalises the association-snapshot artefacts that must
// NOT participate in the differential:
//   - legacy stores embedded copies of associated records whose Users field
//     is whatever the writing path happened to leave (json:"-" — invisible to
//     the API); SQL rebuilds associations live.
//   - nil and empty slices are the same observation for association lists.
//
// Content is only elided when empty; a populated association still compares.
func normaliseRoles(in []models.Role) []models.Role {
	out := append([]models.Role(nil), in...)
	for i := range out {
		if len(out[i].Users) == 0 {
			out[i].Users = []models.User{}
		}
		if len(out[i].Claims) == 0 {
			out[i].Claims = nil
		} else {
			out[i].Claims = normaliseClaims(out[i].Claims)
		}
	}
	return out
}

func normaliseClaims(in []models.Claim) []models.Claim {
	out := append([]models.Claim(nil), in...)
	for i := range out {
		if len(out[i].Users) == 0 {
			out[i].Users = []models.User{}
		}
	}
	return out
}
