package data

import (
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (j *JsonDatabase) GetUsers() ([]models.User, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	return j.data.Users, nil
}

func (j *JsonDatabase) GetUser(idOrEmail string) (*models.User, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}

	for _, user := range j.data.Users {
		if strings.EqualFold(user.ID, idOrEmail) || strings.EqualFold(user.Email, idOrEmail) || strings.EqualFold(user.Username, idOrEmail) {
			return &user, nil
		}
	}

	return nil, fmt.Errorf("User not found")
}

func (j *JsonDatabase) CreateUser(user *models.User) (*models.User, error) {
	if !j.IsConnected() {
		return nil, errors.New("the database is not connected")
	}
	if user.ID == "" {
		user.ID = helpers.GenerateId()
	}
	if user.Email == "" {
		return nil, errors.New("User email cannot be empty")
	}
	if user.Username == "" {
		return nil, errors.New("User username cannot be empty")
	}
	if user.Name == "" {
		return nil, errors.New("User Name cannot be empty")
	}

	if u, _ := j.GetUser(user.ID); u != nil {
		return nil, fmt.Errorf("User already exists")
	}

	if u, _ := j.GetUser(user.Email); u != nil {
		return nil, fmt.Errorf("User already exists")
	}

	if u, _ := j.GetUser(user.Username); u != nil {
		return nil, fmt.Errorf("User already exists")
	}

	if len(user.Roles) == 0 {
		userRole, err := j.GetRole(constants.USER_ROLE)
		if err != nil {
			return nil, err
		}
		user.Roles = []models.UserRole{
			*userRole,
		}
	}

	if len(user.Claims) == 0 {
		readonlyClaim, err := j.GetClaim(constants.READ_ONLY_CLAIM)
		if err != nil {
			return nil, err
		}
		user.Claims = []models.UserClaim{
			*readonlyClaim,
		}
	}

	// Hash the password with SHA-256
	user.Password = helpers.Sha256Hash(user.Password)

	user.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	user.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	j.data.Users = append(j.data.Users, *user)
	j.save()

	return user, nil
}

func (j *JsonDatabase) UpdateUser(key models.User) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	for i, user := range j.data.Users {
		if user.ID == key.ID {
			if key.Name != "" {
				j.data.Users[i].Name = key.Name
			}
			if key.Password != "" {
				j.data.Users[i].Password = helpers.Sha256Hash(key.Password)
			}

			j.data.Users[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			j.save()
			return nil
		}
	}

	return errors.New("User not found")
}

func (j *JsonDatabase) UpdateRootPassword(newPassword string) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	for i, user := range j.data.Users {
		if user.Email == "root@localhost" {
			j.data.Users[i].Password = helpers.Sha256Hash(newPassword)
			j.data.Users[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			j.save()
			return nil
		}
	}

	return errors.New("User not found")
}

func (j *JsonDatabase) RemoveUser(id string) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if id == "" {
		return nil
	}

	for i, user := range j.data.Users {
		if user.ID == id {
			// Remove the key from the slice
			j.data.Users = append(j.data.Users[:i], j.data.Users[i+1:]...)
			j.save()
			return nil
		}
	}

	return errors.New("User not found")
}

func (j *JsonDatabase) AddRoleToUser(userId string, roleId string) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if userId == "" {
		return errors.New("User ID cannot be empty")
	}

	if roleId == "" {
		return errors.New("Role ID cannot be empty")
	}

	user, err := j.GetUser(userId)
	if err != nil {
		return err
	}

	role, err := j.GetRole(roleId)
	if err != nil {
		return err
	}

	for _, r := range user.Roles {
		if r.ID == role.ID {
			return nil
		}
	}

	for i, c := range j.data.Users {
		if c.ID == userId {
			j.data.Users[i].Roles = append(j.data.Users[i].Roles, *role)
			j.save()
		}
	}

	return nil
}

func (j *JsonDatabase) RemoveRoleFromUser(userId string, roleId string) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if userId == "" {
		return errors.New("User ID cannot be empty")
	}

	if roleId == "" {
		return errors.New("Role ID cannot be empty")
	}

	user, err := j.GetUser(userId)
	if err != nil {
		return err
	}

	role, err := j.GetRole(roleId)
	if err != nil {
		return err
	}

	for i, c := range j.data.Users {
		if c.ID == userId {
			for e, r := range user.Roles {
				if r.ID == role.ID {
					j.data.Users[i].Roles = append(j.data.Users[i].Roles[:e], j.data.Users[i].Roles[e+1:]...)
					j.save()
					return nil
				}
			}
		}
	}

	return nil
}

func (j *JsonDatabase) AddClaimToUser(userId string, claimId string) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if userId == "" {
		return errors.New("User ID cannot be empty")
	}

	if claimId == "" {
		return errors.New("Claim ID cannot be empty")
	}

	user, err := j.GetUser(userId)
	if err != nil {
		return err
	}

	claim, err := j.GetClaim(claimId)
	if err != nil {
		return err
	}

	for _, c := range user.Claims {
		if c.ID == claim.ID {
			return nil
		}
	}

	for i, c := range j.data.Users {
		if c.ID == userId {
			j.data.Users[i].Claims = append(j.data.Users[i].Claims, *claim)
			j.save()
		}
	}

	return nil
}

func (j *JsonDatabase) RemoveClaimFromUser(userId string, claimId string) error {
	if !j.IsConnected() {
		return errors.New("the database is not connected")
	}

	if userId == "" {
		return errors.New("User ID cannot be empty")
	}

	if claimId == "" {
		return errors.New("Claim ID cannot be empty")
	}

	user, err := j.GetUser(userId)
	if err != nil {
		return err
	}

	claim, err := j.GetClaim(claimId)
	if err != nil {
		return err
	}

	for i, c := range j.data.Users {
		if c.ID == userId {
			for e, r := range user.Roles {
				if r.ID == claim.ID {
					j.data.Users[i].Claims = append(j.data.Users[i].Claims[:e], j.data.Users[i].Claims[e+1:]...)
					j.save()
					return nil
				}
			}
		}
	}

	return nil
}
