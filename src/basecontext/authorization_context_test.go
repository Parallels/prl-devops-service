package basecontext

// import (
// 	"context"
// 	"testing"

// 	"github.com/Parallels/prl-devops-service/constants"
// 	"github.com/Parallels/prl-devops-service/models"
// 	"github.com/stretchr/testify/assert"
// )

// func TestCloneAuthorizationContext(t *testing.T) {
// 	t.Run("Clone authorization context with nil base context", func(t *testing.T) {
// 		// Reset the baseAuthorizationCtx to nil
// 		baseAuthorizationCtx = nil

// 		// Call the CloneAuthorizationContext function
// 		result := CloneAuthorizationContext()

// 		// Assert that the result is not nil
// 		assert.NotNil(t, result)

// 		// Assert that the result is a new context with default values
// 		assert.Equal(t, "", result.Issuer)
// 		assert.Equal(t, "", result.Scope)
// 		assert.Empty(t, result.Audiences)
// 		assert.Equal(t, "", result.BaseUrl)
// 		assert.False(t, result.IsAuthorized)
// 		assert.Equal(t, "", result.RequestId)
// 		assert.Equal(t, "", result.AuthorizedBy)
// 		assert.Nil(t, result.User)
// 	})

// 	t.Run("Clone authorization context with non-nil base context", func(t *testing.T) {
// 		// Create a baseAuthorizationCtx with some values
// 		baseAuthorizationCtx = &AuthorizationContext{
// 			Issuer:       "example",
// 			Scope:        "read write",
// 			Audiences:    []string{},
// 			BaseUrl:      "https://example.com",
// 			IsAuthorized: false,
// 			RequestId:    "",
// 			AuthorizedBy: "",
// 			User:         nil,
// 		}

// 		// Call the CloneAuthorizationContext function
// 		result := CloneAuthorizationContext()

// 		// Assert that the result is not nil
// 		assert.NotNil(t, result)

// 		// Assert that the result is a new context with the same values as the base context
// 		assert.Equal(t, baseAuthorizationCtx.Issuer, result.Issuer)
// 		assert.Equal(t, baseAuthorizationCtx.Scope, result.Scope)
// 		assert.Equal(t, baseAuthorizationCtx.Audiences, result.Audiences)
// 		assert.Equal(t, baseAuthorizationCtx.BaseUrl, result.BaseUrl)
// 		assert.Equal(t, baseAuthorizationCtx.IsAuthorized, result.IsAuthorized)
// 		assert.Equal(t, baseAuthorizationCtx.RequestId, result.RequestId)
// 		assert.Equal(t, baseAuthorizationCtx.AuthorizedBy, result.AuthorizedBy)
// 		assert.Equal(t, baseAuthorizationCtx.User, result.User)
// 	})
// }

// func TestGetAuthorizationContext(t *testing.T) {
// 	t.Run("Get authorization context with nil context", func(t *testing.T) {
// 		// Call the GetAuthorizationContext function with nil context
// 		result := GetAuthorizationContext(context.TODO())

// 		// Assert that the result is nil
// 		assert.Nil(t, result)
// 	})

// 	t.Run("Get authorization context with non-nil context", func(t *testing.T) {
// 		// Create a context with an authorization context value
// 		authContext := &AuthorizationContext{}
// 		ctx := context.WithValue(context.Background(), constants.AUTHORIZATION_CONTEXT_KEY, authContext)

// 		// Call the GetAuthorizationContext function with the context
// 		result := GetAuthorizationContext(ctx)

// 		// Assert that the result is the same as the authorization context
// 		assert.Equal(t, authContext, result)
// 	})
// }

// func TestAuthorizationContext_UserHasClaim(t *testing.T) {
// 	// Create a test user with claims
// 	user := &models.ApiUser{
// 		Claims: []string{"claim1", "claim2", "claim3"},
// 	}

// 	// Create an authorization context with the test user
// 	authContext := &AuthorizationContext{
// 		User: user,
// 	}

// 	t.Run("User has the claim", func(t *testing.T) {
// 		// Call the UserHasClaim method with an existing claim
// 		result := authContext.UserHasClaim("claim2")

// 		// Assert that the result is true
// 		assert.True(t, result)
// 	})

// 	t.Run("User does not have the claim", func(t *testing.T) {
// 		// Call the UserHasClaim method with a non-existing claim
// 		result := authContext.UserHasClaim("claim4")

// 		// Assert that the result is false
// 		assert.False(t, result)
// 	})

// 	t.Run("User is nil", func(t *testing.T) {
// 		// Create an authorization context with a nil user
// 		authContext := &AuthorizationContext{
// 			User: nil,
// 		}

// 		// Call the UserHasClaim method with a claim
// 		result := authContext.UserHasClaim("claim1")

// 		// Assert that the result is false
// 		assert.False(t, result)
// 	})
// }

// func TestAuthorizationContext_IsUserInRoles(t *testing.T) {
// 	t.Run("User is in roles", func(t *testing.T) {
// 		// Create a test user with roles
// 		user := &models.ApiUser{
// 			Roles: []string{"role1", "role2", "role3"},
// 		}

// 		// Create an authorization context with the test user
// 		authContext := &AuthorizationContext{
// 			User: user,
// 		}

// 		// Define the roles to check
// 		roles := []string{"role2", "role4"}

// 		// Call the IsUserInRoles method
// 		result := authContext.IsUserInRoles(roles)

// 		// Assert that the result is true
// 		assert.True(t, result)
// 	})

// 	t.Run("User is not in roles", func(t *testing.T) {
// 		// Create a test user with roles
// 		user := &models.ApiUser{
// 			Roles: []string{"role1", "role2", "role3"},
// 		}

// 		// Create an authorization context with the test user
// 		authContext := &AuthorizationContext{
// 			User: user,
// 		}

// 		// Define the roles to check
// 		roles := []string{"role4", "role5"}

// 		// Call the IsUserInRoles method
// 		result := authContext.IsUserInRoles(roles)

// 		// Assert that the result is false
// 		assert.False(t, result)
// 	})

// 	t.Run("User is nil", func(t *testing.T) {
// 		// Create an authorization context with a nil user
// 		authContext := &AuthorizationContext{
// 			User: nil,
// 		}

// 		// Define the roles to check
// 		roles := []string{"role1", "role2"}

// 		// Call the IsUserInRoles method
// 		result := authContext.IsUserInRoles(roles)

// 		// Assert that the result is false
// 		assert.False(t, result)
// 	})
// }

// func TestGetBaseContext(t *testing.T) {
// 	t.Run("Get base context when it is nil", func(t *testing.T) {
// 		// Reset the baseAuthorizationCtx to nil
// 		baseAuthorizationCtx = nil

// 		// Call the GetBaseContext function
// 		result := GetBaseContext()

// 		// Assert that the result is not nil
// 		assert.NotNil(t, result)

// 		// Assert that the result is the same as the initialized authorization context
// 		assert.Equal(t, InitAuthorizationContext(), result)
// 	})

// 	t.Run("Get base context when it is not nil", func(t *testing.T) {
// 		// Create a baseAuthorizationCtx with some values
// 		baseAuthorizationCtx = &AuthorizationContext{
// 			Issuer:       "example",
// 			Scope:        "read write",
// 			Audiences:    []string{},
// 			BaseUrl:      "https://example.com",
// 			IsAuthorized: false,
// 			RequestId:    "",
// 			AuthorizedBy: "",
// 			User:         nil,
// 		}

// 		// Call the GetBaseContext function
// 		result := GetBaseContext()

// 		// Assert that the result is not nil
// 		assert.NotNil(t, result)

// 		// Assert that the result is the same as the baseAuthorizationCtx
// 		assert.Equal(t, baseAuthorizationCtx, result)
// 	})
// }

// func TestAuthorizationContext_IsUserInRole(t *testing.T) {
// 	t.Run("User is in role", func(t *testing.T) {
// 		// Create a test user with roles
// 		user := &models.ApiUser{
// 			Roles: []string{"role1", "role2", "role3"},
// 		}

// 		// Create an authorization context with the test user
// 		authContext := &AuthorizationContext{
// 			User: user,
// 		}

// 		// Call the IsUserInRole method with an existing role
// 		result := authContext.IsUserInRole("role2")

// 		// Assert that the result is true
// 		assert.True(t, result)
// 	})

// 	t.Run("User is not in role", func(t *testing.T) {
// 		// Create a test user with roles
// 		user := &models.ApiUser{
// 			Roles: []string{"role1", "role2", "role3"},
// 		}

// 		// Create an authorization context with the test user
// 		authContext := &AuthorizationContext{
// 			User: user,
// 		}

// 		// Call the IsUserInRole method with a non-existing role
// 		result := authContext.IsUserInRole("role4")

// 		// Assert that the result is false
// 		assert.False(t, result)
// 	})

// 	t.Run("User is nil", func(t *testing.T) {
// 		// Create an authorization context with a nil user
// 		authContext := &AuthorizationContext{
// 			User: nil,
// 		}

// 		// Call the IsUserInRole method with a role
// 		result := authContext.IsUserInRole("role1")

// 		// Assert that the result is false
// 		assert.False(t, result)
// 	})
// }
