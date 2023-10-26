package data

import (
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"errors"
	"fmt"
	"strings"
)

func (j *JsonDatabase) GetClaims() ([]models.Claim, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	return j.data.Claims, nil
}

func (j *JsonDatabase) GetClaim(idOrName string) (*models.Claim, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	for _, claim := range j.data.Claims {
		if strings.EqualFold(claim.ID, idOrName) || strings.EqualFold(claim.Name, idOrName) {
			return &claim, nil
		}
	}

	return nil, fmt.Errorf("claim not found")
}

func (j *JsonDatabase) CreateClaim(claim *models.Claim) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if claim.ID == "" {
		claim.ID = helpers.GenerateId()
	}

	if claim.Name == "" {
		return errors.New("claim does not contain a name")
	}

	if u, _ := j.GetUser(claim.ID); u != nil {
		return fmt.Errorf("claim %s already exists with ID %s", claim.Name, claim.ID)
	}

	if u, _ := j.GetUser(claim.Name); u != nil {
		return fmt.Errorf("claim %s already exists with ID %s", claim.Name, claim.ID)
	}

	j.data.Claims = append(j.data.Claims, *claim)
	j.save()
	return nil
}
