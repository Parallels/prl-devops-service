package stores

import (
	"fmt"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/data/models"
	"github.com/Parallels/prl-devops-service/database/interfaces"
	apperrors "github.com/Parallels/prl-devops-service/errors"
)

// Compile-time interface compliance check
var _ interfaces.AuthDomain = (*AuthService)(nil)

// AuthService handles authentication and authorization domain operations with business logic
// For simple CRUD operations, use db.Stores() directly instead of adding pass-through methods here
type AuthService struct {
	userStore  UserDataStoreInterface
	roleStore  RoleDataStoreInterface
	claimStore ClaimDataStoreInterface
}

// NewAuthService creates a new auth domain service
func NewAuthService(
	userStore UserDataStoreInterface,
	roleStore RoleDataStoreInterface,
	claimStore ClaimDataStoreInterface,
) *AuthService {
	return &AuthService{
		userStore:  userStore,
		roleStore:  roleStore,
		claimStore: claimStore,
	}
}

// toBaseContext converts ApiContext to BaseContext
func toBaseContext(ctx basecontext.ApiContext) *basecontext.BaseContext {
	baseCtx, ok := ctx.(*basecontext.BaseContext)
	if !ok {
		baseCtx = basecontext.NewBaseContextFromContext(ctx.Context())
	}
	return baseCtx
}

// GetUser retrieves a user by ID or username with smart fallback logic
// This method adds value by trying multiple lookup strategies
func (s *AuthService) GetUser(ctx basecontext.ApiContext, idOrEmail string) (*models.User, *apperrors.Diagnostics) {
	baseCtx := toBaseContext(ctx)

	// Try by ID first
	user, diag := s.userStore.GetUserByID(*baseCtx, idOrEmail)
	if diag != nil && !diag.HasErrors() && user != nil {
		return user, nil
	}

	// Try by username
	user, diag = s.userStore.GetUserByUsername(*baseCtx, idOrEmail)
	if diag != nil && diag.HasErrors() {
		return nil, diag
	}
	if user == nil {
		diag = apperrors.NewDiagnostics("get_user")
		diag.AddError("user_not_found", fmt.Sprintf("user not found: %s", idOrEmail), "auth_service", nil)
		return nil, diag
	}
	return user, nil
}

// Future business logic methods go here:
// - RegisterUser(ctx, email, password) - create user + send welcome email + audit log
// - LoginUser(ctx, username, password) - validate + create session + log access
// - PromoteToAdmin(ctx, userId) - assign admin role + claims + send notification
// - etc.
