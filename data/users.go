package data

import (
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/helpers"
	"errors"
	"strings"
	"time"
)

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrUserAlreadyExists         = errors.New("user already exists")
	ErrUserEmailCannotBeEmpty    = errors.New("user email cannot be empty")
	ErrUserUsernameCannotBeEmpty = errors.New("user username cannot be empty")
	ErrUserNameCannotBeEmpty     = errors.New("user Name cannot be empty")
	ErrUserIDCannotBeEmpty       = errors.New("user ID cannot be empty")
	ErrRoleIDCannotBeEmpty       = errors.New("role ID cannot be empty")
	ErrClaimIDCannotBeEmpty      = errors.New("claim ID cannot be empty")
)

func (j *JsonDatabase) GetUsers() ([]models.User, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	return j.data.Users, nil
}

func (j *JsonDatabase) GetUser(idOrEmail string) (*models.User, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	for _, user := range j.data.Users {
		if strings.EqualFold(user.ID, idOrEmail) || strings.EqualFold(user.Email, idOrEmail) || strings.EqualFold(user.Username, idOrEmail) {
			return &user, nil
		}
	}

	return nil, ErrUserNotFound
}

func (j *JsonDatabase) CreateUser(user *models.User) (*models.User, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}
	if user.ID == "" {
		user.ID = helpers.GenerateId()
	}
	if user.Email == "" {
		return nil, ErrUserEmailCannotBeEmpty
	}
	if user.Username == "" {
		return nil, ErrUserUsernameCannotBeEmpty
	}
	if user.Name == "" {
		return nil, ErrUserNameCannotBeEmpty
	}

	if u, _ := j.GetUser(user.ID); u != nil {
		return nil, ErrUserAlreadyExists
	}

	if u, _ := j.GetUser(user.Email); u != nil {
		return nil, ErrUserAlreadyExists
	}

	if u, _ := j.GetUser(user.Username); u != nil {
		return nil, ErrUserAlreadyExists
	}

	if len(user.Roles) == 0 {
		userRole, err := j.GetRole(constants.USER_ROLE)
		if err != nil {
			return nil, err
		}
		user.Roles = []models.Role{
			*userRole,
		}
	}

	if len(user.Claims) == 0 {
		readonlyClaim, err := j.GetClaim(constants.READ_ONLY_CLAIM)
		if err != nil {
			return nil, err
		}
		user.Claims = []models.Claim{
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
		return ErrDatabaseNotConnected
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

	return ErrUserNotFound
}

func (j *JsonDatabase) UpdateRootPassword(newPassword string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, user := range j.data.Users {
		if user.Email == "root@localhost" {
			j.data.Users[i].Password = helpers.Sha256Hash(newPassword)
			j.data.Users[i].UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
			j.save()
			return nil
		}
	}

	return ErrUserNotFound
}

func (j *JsonDatabase) RemoveUser(id string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
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

	return ErrUserNotFound
}

func (j *JsonDatabase) AddRoleToUser(userId string, roleId string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if userId == "" {
		return ErrUserIDCannotBeEmpty
	}

	if roleId == "" {
		return ErrRoleIDCannotBeEmpty
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
		return ErrDatabaseNotConnected
	}

	if userId == "" {
		return ErrUserIDCannotBeEmpty
	}

	if roleId == "" {
		return ErrRoleIDCannotBeEmpty
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
		return ErrDatabaseNotConnected
	}

	if userId == "" {
		return ErrUserIDCannotBeEmpty
	}

	if claimId == "" {
		return ErrClaimIDCannotBeEmpty
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
		return ErrDatabaseNotConnected
	}

	if userId == "" {
		return ErrUserIDCannotBeEmpty
	}

	if claimId == "" {
		return ErrClaimIDCannotBeEmpty
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
