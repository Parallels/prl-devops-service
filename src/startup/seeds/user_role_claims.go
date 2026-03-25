package seeds

// SeedUsersMissingClaimsByRole is retained for backward compatibility but is
// now a no-op. Role-based claims are no longer stamped directly onto users;
// they are attached to the role itself and resolved at runtime via
// ComputeEffectiveClaims. See SeedDefaultRoleClaims for the replacement logic.
func SeedUsersMissingClaimsByRole() error {
	return nil
}
