package data

import (
	"Parallels/pd-api-service/basecontext"
	"Parallels/pd-api-service/constants"
	"Parallels/pd-api-service/data/models"
	"Parallels/pd-api-service/errors"
	"Parallels/pd-api-service/helpers"
	"strings"
)

var (
	ErrUserNotFound              = errors.NewWithCode("user not found", 404)
	ErrUserAlreadyExists         = errors.NewWithCode("user already exists", 400)
	ErrUserEmailCannotBeEmpty    = errors.NewWithCode("user email cannot be empty", 400)
	ErrUserUsernameCannotBeEmpty = errors.NewWithCode("user username cannot be empty", 400)
	ErrUserNameCannotBeEmpty     = errors.NewWithCode("user name cannot be empty", 400)
	ErrUserIDCannotBeEmpty       = errors.NewWithCode("user ID cannot be empty", 400)
	ErrRoleIDCannotBeEmpty       = errors.NewWithCode("role ID cannot be empty", 400)
	ErrClaimIDCannotBeEmpty      = errors.NewWithCode("claim ID cannot be empty", 400)
	ErrCannotUpdateRootUser      = errors.NewWithCode("cannot update root user", 400)
	ErrCannotRemoveRootUser      = errors.NewWithCode("cannot remove root user", 400)
	ErrUserAlreadyContainsRole   = errors.NewWithCode("user already contains role", 400)
	ErrUserAlreadyContainsClaim  = errors.NewWithCode("user already contains claim", 400)
)

func (j *JsonDatabase) GetUsers(ctx basecontext.ApiContext, filter string) ([]models.User, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	dbFilter, err := ParseFilter(filter)
	if err != nil {
		return nil, err
	}

	filteredData, err := FilterByProperty(j.data.Users, dbFilter)
	if err != nil {
		return nil, err
	}

	return filteredData, nil
}

func (j *JsonDatabase) GetUser(ctx basecontext.ApiContext, idOrEmail string) (*models.User, error) {
	if !j.IsConnected() {
		return nil, ErrDatabaseNotConnected
	}

	users, err := j.GetUsers(ctx, "")
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if strings.EqualFold(user.ID, idOrEmail) || strings.EqualFold(user.Email, idOrEmail) || strings.EqualFold(user.Username, idOrEmail) {
			return &user, nil
		}
	}

	return nil, ErrUserNotFound
}

func (j *JsonDatabase) CreateUser(ctx basecontext.ApiContext, user models.User) (*models.User, error) {
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

	if u, _ := j.GetUser(ctx, user.ID); u != nil {
		return nil, ErrUserAlreadyExists
	}

	if u, _ := j.GetUser(ctx, user.Email); u != nil {
		return nil, ErrUserAlreadyExists
	}

	if u, _ := j.GetUser(ctx, user.Username); u != nil {
		return nil, ErrUserAlreadyExists
	}

	if len(user.Roles) == 0 {
		for _, role := range constants.DefaultRoles {
			dbRole, err := j.GetRole(ctx, role)
			if err != nil {
				return nil, err
			}
			user.Roles = append(user.Roles, *dbRole)
		}
	} else {
		for _, role := range user.Roles {
			_, err := j.GetRole(ctx, role.ID)
			if err != nil {
				return nil, errors.NewWithCodef(400, "role %s does not exist", role.Name)
			}
		}
	}

	if len(user.Claims) == 0 {
		for _, claim := range constants.DefaultClaims {
			dbClaim, err := j.GetClaim(ctx, claim)
			if err != nil {
				return nil, err
			}
			user.Claims = append(user.Claims, *dbClaim)
		}
	} else {
		for _, claim := range user.Claims {
			_, err := j.GetClaim(ctx, claim.ID)
			if err != nil {
				return nil, errors.NewWithCodef(400, "claim %s does not exist", claim.Name)
			}
		}
	}

	// Hash the password with SHA-256
	user.Password = helpers.Sha256Hash(user.Password)

	user.UpdatedAt = helpers.GetUtcCurrentDateTime()
	user.CreatedAt = helpers.GetUtcCurrentDateTime()
	j.data.Users = append(j.data.Users, user)
	j.Save(ctx)

	return &user, nil
}

func (j *JsonDatabase) UpdateUser(ctx basecontext.ApiContext, key models.User) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if key.Name == "root" || key.Username == "root" || key.Email == "root@localhost" {
		return ErrCannotUpdateRootUser
	}

	for i, user := range j.data.Users {
		if user.ID == key.ID {
			if key.Name != "" {
				j.data.Users[i].Name = key.Name
			}
			if key.Password != "" {
				j.data.Users[i].Password = helpers.Sha256Hash(key.Password)
			}

			j.data.Users[i].UpdatedAt = helpers.GetUtcCurrentDateTime()
			j.Save(ctx)
			return nil
		}
	}

	return ErrUserNotFound
}

func (j *JsonDatabase) UpdateRootPassword(ctx basecontext.ApiContext, newPassword string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	for i, user := range j.data.Users {
		if user.Email == "root@localhost" {
			j.data.Users[i].Password = helpers.Sha256Hash(newPassword)
			j.data.Users[i].UpdatedAt = helpers.GetUtcCurrentDateTime()
			j.Save(ctx)
			return nil
		}
	}

	return ErrUserNotFound
}

func (j *JsonDatabase) DeleteUser(ctx basecontext.ApiContext, id string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if id == "" {
		return nil
	}

	for i, user := range j.data.Users {
		if user.ID == id {
			if user.Name == "root" || user.Username == "root" || user.Email == "root@localhost" {
				return ErrCannotUpdateRootUser
			}

			j.data.Users = append(j.data.Users[:i], j.data.Users[i+1:]...)
			j.Save(ctx)
			return nil
		}
	}

	return ErrUserNotFound
}

func (j *JsonDatabase) AddRoleToUser(ctx basecontext.ApiContext, userId string, roleId string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if userId == "" {
		return ErrUserIDCannotBeEmpty
	}

	if roleId == "" {
		return ErrRoleIDCannotBeEmpty
	}

	user, err := j.GetUser(ctx, userId)
	if err != nil {
		return err
	}

	role, err := j.GetRole(ctx, roleId)
	if err != nil {
		return err
	}

	for _, r := range user.Roles {
		if r.ID == role.ID {
			return ErrUserAlreadyContainsRole
		}
	}

	for i, c := range j.data.Users {
		if c.ID == userId {
			j.data.Users[i].Roles = append(j.data.Users[i].Roles, *role)
			j.Save(ctx)
			return nil
		}
	}

	return ErrUserNotFound
}

func (j *JsonDatabase) RemoveRoleFromUser(ctx basecontext.ApiContext, userId string, roleId string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if userId == "" {
		return ErrUserIDCannotBeEmpty
	}

	if roleId == "" {
		return ErrRoleIDCannotBeEmpty
	}

	user, err := j.GetUser(ctx, userId)
	if err != nil {
		return err
	}

	role, err := j.GetRole(ctx, roleId)
	if err != nil {
		return err
	}

	for i, c := range j.data.Users {
		if c.ID == userId {
			for e, r := range user.Roles {
				if r.ID == role.ID {
					j.data.Users[i].Roles = append(j.data.Users[i].Roles[:e], j.data.Users[i].Roles[e+1:]...)
					j.Save(ctx)
					return nil
				}
			}
		}
	}

	return ErrRoleNotFound
}

func (j *JsonDatabase) AddClaimToUser(ctx basecontext.ApiContext, userId string, claimId string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if userId == "" {
		return ErrUserIDCannotBeEmpty
	}

	if claimId == "" {
		return ErrClaimIDCannotBeEmpty
	}

	user, err := j.GetUser(ctx, userId)
	if err != nil {
		return err
	}

	claim, err := j.GetClaim(ctx, claimId)
	if err != nil {
		return err
	}

	for _, r := range user.Claims {
		if r.ID == claim.ID {
			return ErrUserAlreadyContainsClaim
		}
	}

	for i, c := range j.data.Users {
		if c.ID == userId {
			j.data.Users[i].Claims = append(j.data.Users[i].Claims, *claim)
			j.Save(ctx)
			return nil
		}
	}

	return ErrClaimNotFound
}

func (j *JsonDatabase) RemoveClaimFromUser(ctx basecontext.ApiContext, userId string, claimId string) error {
	if !j.IsConnected() {
		return ErrDatabaseNotConnected
	}

	if userId == "" {
		return ErrUserIDCannotBeEmpty
	}

	if claimId == "" {
		return ErrClaimIDCannotBeEmpty
	}

	user, err := j.GetUser(ctx, userId)
	if err != nil {
		return err
	}

	claim, err := j.GetClaim(ctx, claimId)
	if err != nil {
		return err
	}

	for i, c := range j.data.Users {
		if c.ID == userId {
			for e, r := range user.Claims {
				if r.ID == claim.ID {
					j.data.Users[i].Claims = append(j.data.Users[i].Claims[:e], j.data.Users[i].Claims[e+1:]...)
					j.Save(ctx)
					return nil
				}
			}
		}
	}

	return ErrClaimNotFound
}
