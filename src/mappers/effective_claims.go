package mappers

import (
	data_models "github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/models"
)

// ComputeEffectiveClaims merges a user's direct claims with the claims inherited
// from each of their roles. The result is deduplicated: if a claim exists both
// directly and via a role, it appears once with IsInherited=false (direct wins).
// If the same claim appears in multiple roles, the first role encountered wins
// for the SourceRole field.
func ComputeEffectiveClaims(user data_models.User) []models.UserClaimResponse {
	effective := make([]models.UserClaimResponse, 0)
	seen := make(map[string]bool)

	// Direct claims first — they take precedence over inherited ones.
	for _, c := range user.Claims {
		id := c.ID
		if seen[id] {
			continue
		}
		seen[id] = true
		effective = append(effective, models.UserClaimResponse{
			ID:          c.ID,
			Name:        c.Name,
			IsInherited: false,
		})
	}

	// Role-inherited claims.
	for _, role := range user.Roles {
		for _, c := range role.Claims {
			id := c.ID
			if seen[id] {
				continue
			}
			seen[id] = true
			effective = append(effective, models.UserClaimResponse{
				ID:          c.ID,
				Name:        c.Name,
				IsInherited: true,
				SourceRole:  role.ID,
			})
		}
	}

	return effective
}

// ComputeEffectiveClaimIDs returns just the claim IDs from ComputeEffectiveClaims,
// suitable for middleware claim-checking and JWT generation.
func ComputeEffectiveClaimIDs(user data_models.User) []string {
	claims := ComputeEffectiveClaims(user)
	ids := make([]string, 0, len(claims))
	for _, c := range claims {
		ids = append(ids, c.ID)
	}
	return ids
}
