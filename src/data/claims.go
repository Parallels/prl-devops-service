package data

import (
	"strings"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/errors"
	"github.com/Parallels/prl-devops-service/helpers"
)

var (
	ErrClaimEmptyNameOrId  = errors.NewWithCode("no claim specified", 500)
	ErrClaimEmptyName      = errors.NewWithCode("claim name cannot be empty", 500)
	ErrClaimNotFound       = errors.NewWithCode("claim not found", 404)
	ErrRemoveInternalClaim = errors.NewWithCode("claim is internal and cannot be removed", 400)
	ErrUpdateInternalClaim = errors.NewWithCode("claim is internal and cannot be updated", 400)
)

func (j *JsonDatabase) GetClaims(ctx basecontext.ApiContext, filter string) ([]models.Claim, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(j.data.Claims, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

func (j *JsonDatabase) GetClaim(ctx basecontext.ApiContext, idOrName string) (*models.Claim, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	claims, err := j.GetClaims(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, claim := range claims {
		if strings.EqualFold(claim.ID, idOrName) || strings.EqualFold(claim.Name, idOrName) {
			return &claim, nil
		}
	}

	return nil, ErrClaimNotFound
}

func (j *JsonDatabase) CreateClaim(ctx basecontext.ApiContext, claim models.Claim) (*models.Claim, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if claim.Name == "" {
		return nil, ErrClaimEmptyName
	}

	claim.Name = strings.ToUpper(helpers.NormalizeString(claim.Name))
	claim.ID = claim.Name

	if u, _ := j.GetClaim(ctx, claim.ID); u != nil {
		return nil, errors.NewWithCodef(400, "claim %s already exists with ID %s", claim.Name, claim.ID)
	}

	j.data.Claims = append(j.data.Claims, claim)
	if err := j.Save(ctx); err != nil {
		return nil, err
	}

	return &claim, nil
}

func (j *JsonDatabase) DeleteClaim(ctx basecontext.ApiContext, idOrName string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if idOrName == "" {
		return ErrClaimEmptyNameOrId
	}

	for i, claim := range j.data.Claims {
		if strings.EqualFold(claim.ID, idOrName) || strings.EqualFold(claim.Name, idOrName) {
			if claim.Internal && !IsRootUser(ctx) {
				return ErrRemoveInternalClaim
			}

			j.data.Claims = append(j.data.Claims[:i], j.data.Claims[i+1:]...)
			for _, user := range j.data.Users {
				for j, userClaim := range user.Claims {
					if userClaim.ID == claim.ID {
						user.Claims = append(user.Claims[:j], user.Claims[j+1:]...)
					}
				}
			}
			if err := j.Save(ctx); err != nil {
				return err
			}
			return nil
		}
	}

	return ErrClaimNotFound
}

func (j *JsonDatabase) UpdateClaim(ctx basecontext.ApiContext, claim *models.Claim) (*models.Claim, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	if claim.ID == "" {
		return nil, ErrClaimEmptyNameOrId
	}

	for i, c := range j.data.Claims {
		if strings.EqualFold(c.ID, claim.ID) || strings.EqualFold(c.Name, claim.Name) {
			if claim.Internal {
				return nil, ErrUpdateInternalClaim
			}
			oldClaim := j.data.Claims[i]
			claim.ID = strings.ToUpper(helpers.NormalizeString(claim.Name))
			j.data.Claims[i] = *claim
			for _, user := range j.data.Users {
				for j, userClaim := range user.Claims {
					if userClaim.ID == oldClaim.ID {
						user.Claims[j] = *claim
					}
				}
			}
			if err := j.Save(ctx); err != nil {
				return nil, err
			}
			return claim, nil
		}
	}

	return nil, ErrClaimNotFound
}
