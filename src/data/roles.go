package data

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

var (
	ErrRoleEmptyNameOrId        = errors.NewWithCode("no role specified", 500)
	ErrRoleEmptyName            = errors.NewWithCode("role name cannot be empty", 500)
	ErrRoleNotFound             = errors.NewWithCode("role not found", 404)
	ErrRemoveInternalRole       = errors.NewWithCode("role is internal and cannot be removed", 400)
	ErrUpdateInternalRole       = errors.NewWithCode("role is internal and cannot be updated", 400)
	ErrRoleAlreadyContainsClaim = errors.NewWithCode("role already contains claim", 400)
	ErrRoleDoesNotContainClaim  = errors.NewWithCode("role does not contain claim", 404)
)

func (j *JsonDatabase) GetRoles(ctx basecontext.ApiContext, filter string) ([]models.Role, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}
	users, err := j.GetUsers(ctx, "")
	if err != nil {
		return nil, err
	}

	// Add users to roles
	for i, role := range j.data.Roles {
		// reset users
		j.data.Roles[i].Users = make([]models.User, 0)
		for _, user := range users {
			for _, r := range user.Roles {
				if strings.EqualFold(r.ID, role.ID) {
					// adding the user to the role if
					j.data.Roles[i].Users = append(j.data.Roles[i].Users, user)
				}
			}
		}
	}

	filteredData, err := FilterByProperty(j.data.Roles, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

func (j *JsonDatabase) GetRole(ctx basecontext.ApiContext, idOrName string) (*models.Role, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	roles, err := j.GetRoles(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, role := range roles {
		if strings.EqualFold(role.ID, idOrName) || strings.EqualFold(role.Name, idOrName) {
			users, err := j.GetUsers(ctx, "")
			if err != nil {
				return nil, err
			}
			// reset users
			role.Users = make([]models.User, 0)
			for _, user := range users {
				for _, r := range user.Roles {
					if strings.EqualFold(r.ID, role.ID) {
						// adding the user to the role if
						role.Users = append(role.Users, user)
					}
				}
			}
			return &role, nil
		}
	}

	return nil, ErrRoleNotFound
}

func (j *JsonDatabase) CreateRole(ctx basecontext.ApiContext, role models.Role) (*models.Role, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if role.Name == "" {
		return nil, ErrRoleEmptyName
	}

	role.Name = strings.ToUpper(helpers.NormalizeString(role.Name))
	role.ID = role.Name

	if r, _ := j.GetRole(ctx, role.ID); r != nil {
		return nil, errors.NewWithCodef(400, "role %s already exists with ID %s", role.Name, role.ID)
	}

	// Resolve any claims provided against the DB to ensure they exist.
	if len(role.Claims) > 0 {
		resolved := make([]models.Claim, 0, len(role.Claims))
		for _, c := range role.Claims {
			dbClaim, err := j.GetClaim(ctx, c.ID)
			if err != nil {
				return nil, errors.NewWithCodef(400, "claim %s not found", c.ID)
			}
			resolved = append(resolved, *dbClaim)
		}
		role.Claims = resolved
	}

	j.data.Roles = append(j.data.Roles, role)

	return &role, nil
}

func (j *JsonDatabase) AddClaimToRole(ctx basecontext.ApiContext, roleIdOrName string, claimIdOrName string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	dbClaim, err := j.GetClaim(ctx, claimIdOrName)
	if err != nil {
		return ErrClaimNotFound
	}

	for i, role := range j.data.Roles {
		if strings.EqualFold(role.ID, roleIdOrName) || strings.EqualFold(role.Name, roleIdOrName) {
			for _, existing := range role.Claims {
				if strings.EqualFold(existing.ID, dbClaim.ID) {
					return ErrRoleAlreadyContainsClaim
				}
			}
			j.data.Roles[i].Claims = append(j.data.Roles[i].Claims, *dbClaim)
			return nil
		}
	}

	return ErrRoleNotFound
}

func (j *JsonDatabase) RemoveClaimFromRole(ctx basecontext.ApiContext, roleIdOrName string, claimIdOrName string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, role := range j.data.Roles {
		if strings.EqualFold(role.ID, roleIdOrName) || strings.EqualFold(role.Name, roleIdOrName) {
			for k, existing := range j.data.Roles[i].Claims {
				if strings.EqualFold(existing.ID, claimIdOrName) || strings.EqualFold(existing.Name, claimIdOrName) {
					j.data.Roles[i].Claims = append(j.data.Roles[i].Claims[:k], j.data.Roles[i].Claims[k+1:]...)
					return nil
				}
			}
			return ErrRoleDoesNotContainClaim
		}
	}

	return ErrRoleNotFound
}

func (j *JsonDatabase) DeleteRole(ctx basecontext.ApiContext, idOrName string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if idOrName == "" {
		return ErrRoleEmptyNameOrId
	}

	for i, role := range j.data.Roles {
		if strings.EqualFold(role.ID, idOrName) || strings.EqualFold(role.Name, idOrName) {
			if role.Internal && !IsRootUser(ctx) {
				return ErrRemoveInternalRole
			}
			j.data.Roles = append(j.data.Roles[:i], j.data.Roles[i+1:]...)
			return nil
		}
	}

	return ErrRoleNotFound
}

// UpdateRoleDescription sets the Description field on a role without triggering
// the internal-role guard. Used on startup to backfill existing installations.
func (j *JsonDatabase) UpdateRoleDescription(ctx basecontext.ApiContext, idOrName, description string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}
	if idOrName == "" {
		return ErrRoleEmptyNameOrId
	}
	for i, r := range j.data.Roles {
		if strings.EqualFold(r.ID, idOrName) || strings.EqualFold(r.Name, idOrName) {
			j.data.Roles[i].Description = description
			return nil
		}
	}
	return ErrRoleNotFound
}

func (j *JsonDatabase) UpdateRole(ctx basecontext.ApiContext, role *models.Role) (*models.Role, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if role.ID == "" {
		return nil, ErrRoleEmptyNameOrId
	}

	for i, c := range j.data.Roles {
		if strings.EqualFold(c.ID, role.ID) || strings.EqualFold(c.Name, role.Name) {
			if role.Internal {
				return nil, ErrUpdateInternalRole
			}

			j.data.Roles[i] = *role

			return role, nil
		}
	}

	return nil, ErrRoleNotFound
}
