package database

import (
	"context"
	"fmt"

	"github.com/Parallels/prl-devops-service/database/stores"
	"gorm.io/gorm"
)

// StoreRegistry manages all database stores
type StoreRegistry struct {
	db *gorm.DB

	userStore       stores.UserDataStoreInterface
	userConfigStore stores.UserConfigDataStoreInterface
	roleStore       stores.RoleDataStoreInterface
	claimStore      stores.ClaimDataStoreInterface
}

// NewStoreRegistry creates and initializes all stores
func NewStoreRegistry(db *gorm.DB) (*StoreRegistry, error) {
	ctx := context.Background()

	// Initialize user store
	userStore := stores.GetUserDataStoreInstance()
	if err := userStore.Init(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to initialize user store: %w", err)
	}

	// Initialize user config store
	userConfigStore := stores.GetUserConfigDataStoreInstance()
	if err := userConfigStore.Init(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to initialize user config store: %w", err)
	}

	// Initialize role store
	roleStore := stores.GetRoleDataStoreInstance()
	if err := roleStore.Init(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to initialize role store: %w", err)
	}

	// Initialize claim store
	claimStore := stores.GetClaimDataStoreInstance()
	if err := claimStore.Init(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to initialize claim store: %w", err)
	}

	return &StoreRegistry{
		db:              db,
		userStore:       userStore,
		userConfigStore: userConfigStore,
		roleStore:       roleStore,
		claimStore:      claimStore,
	}, nil
}

// User returns the user data store
func (r *StoreRegistry) User() stores.UserDataStoreInterface {
	return r.userStore
}

// UserConfig returns the user config data store
func (r *StoreRegistry) UserConfig() stores.UserConfigDataStoreInterface {
	return r.userConfigStore
}

// Role returns the role data store
func (r *StoreRegistry) Role() stores.RoleDataStoreInterface {
	return r.roleStore
}

// Claim returns the claim data store
func (r *StoreRegistry) Claim() stores.ClaimDataStoreInterface {
	return r.claimStore
}

// DB returns the underlying database connection
func (r *StoreRegistry) DB() *gorm.DB {
	return r.db
}
