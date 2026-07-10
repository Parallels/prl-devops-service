package interfaces

import (
	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	apperrors "github.com/Parallels/prl-devops-service/errors"
)

// AuthDomain defines authentication and authorization operations with business logic
// For simple CRUD operations, use db.Stores() directly
type AuthDomain interface {
	// GetUser retrieves a user by ID or username with smart fallback logic
	GetUser(ctx basecontext.ApiContext, idOrEmail string) (*models.User, *apperrors.Diagnostics)

	// Add future business logic methods here:
	// RegisterUser - create user + send welcome email + audit
	// LoginUser - validate credentials + create session + log
	// PromoteToAdmin - assign role + claims + notify
}
