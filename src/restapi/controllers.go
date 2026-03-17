// Package restapi provides a set of functions to create and register HTTP controllers.
// A controller is a struct that contains information about a specific HTTP endpoint, such as the path, method, and required roles and claims.
// The package also provides functions to serve the registered controllers.
package restapi

import (
	"net/http"
	"strings"

	"github.com/Parallels/prl-devops-service/errors"
	"github.com/cjlapao/common-go/helper/http_helper"
)

// Controller represents a REST API controller with its properties.
type Controller struct {
	listener                 *HttpListener
	path                     string
	Handler                  ControllerHandler
	Version                  HttpVersion
	Method                   HttpControllerMethod
	RequiredRoles            []string
	RequiredClaims           []string
	RoleComparisonOperation  ComparisonOperation
	ClaimComparisonOperation ComparisonOperation
	// ExtraAdapters are injected into the middleware chain just before
	// EndAuthorizationMiddlewareAdapter, allowing per-endpoint auth extensions
	// such as enrollment-token support.
	ExtraAdapters []Adapter
}

// NewController creates a new instance of the Controller struct with default values.
// The returned controller has the GET method, uses the global HTTP listener, and has no required roles or claims.
func NewController() *Controller {
	controller := &Controller{
		Method:                   GET,
		listener:                 globalHttpListener,
		RequiredRoles:            make([]string, 0),
		RequiredClaims:           make([]string, 0),
		RoleComparisonOperation:  ComparisonOperationAnd,
		ClaimComparisonOperation: ComparisonOperationAnd,
	}

	return controller
}

func (c *Controller) WithPath(path string) *Controller {
	c.path = path
	return c
}

func (c *Controller) WithVersion(versionPath string) *Controller {
	if c.listener == nil {
		return c
	}

	for _, version := range c.listener.Versions {
		listenerVersionPath := http_helper.JoinUrl(versionPath)
		path := http_helper.JoinUrl(version.Path)
		if strings.EqualFold(listenerVersionPath, path) {
			c.Version = version
			break
		}
	}

	return c
}

func (c *Controller) WithMethod(method HttpControllerMethod) *Controller {
	c.Method = method
	return c
}

func (c *Controller) WithRequiredRole(role string) *Controller {
	for _, r := range c.RequiredRoles {
		if strings.EqualFold(r, role) {
			return c
		}
	}

	c.RequiredRoles = append(c.RequiredRoles, role)
	return c
}

func (c *Controller) WithRequiredClaim(claim string) *Controller {
	for _, r := range c.RequiredClaims {
		if strings.EqualFold(r, claim) {
			return c
		}
	}

	c.RequiredClaims = append(c.RequiredClaims, claim)
	return c
}

func (c *Controller) WithRoleComparisonOperation(operation ComparisonOperation) *Controller {
	c.RoleComparisonOperation = normalizeComparisonOperation(operation)
	return c
}

func (c *Controller) WithClaimComparisonOperation(operation ComparisonOperation) *Controller {
	c.ClaimComparisonOperation = normalizeComparisonOperation(operation)
	return c
}

func (c *Controller) WithComparisonOperations(roleOperation ComparisonOperation, claimOperation ComparisonOperation) *Controller {
	c.RoleComparisonOperation = normalizeComparisonOperation(roleOperation)
	c.ClaimComparisonOperation = normalizeComparisonOperation(claimOperation)
	return c
}

func (c *Controller) WithAndRoles() *Controller {
	return c.WithRoleComparisonOperation(ComparisonOperationAnd)
}

func (c *Controller) WithOrRoles() *Controller {
	return c.WithRoleComparisonOperation(ComparisonOperationOr)
}

func (c *Controller) WithAndClaims() *Controller {
	return c.WithClaimComparisonOperation(ComparisonOperationAnd)
}

func (c *Controller) WithOrClaims() *Controller {
	return c.WithClaimComparisonOperation(ComparisonOperationOr)
}

func (c *Controller) WithHandler(handler ControllerHandler) *Controller {
	c.Handler = handler
	return c
}

func (c *Controller) WithExtraAdapter(adapter Adapter) *Controller {
	c.ExtraAdapters = append(c.ExtraAdapters, adapter)
	return c
}

func (c *Controller) Register() *Controller {
	for _, controller := range c.listener.Controllers {
		if strings.EqualFold(controller.Path(), c.Path()) && controller.Method == c.Method {
			return c
		}
	}

	c.listener.Controllers = append(c.listener.Controllers, c)
	return c
}

func (c *Controller) Serve() error {
	if c.listener == nil {
		return errors.NewWithCode("listener not found", http.StatusInternalServerError)
	}

	serveAuthorized := func(path string) {
		if len(c.ExtraAdapters) > 0 {
			c.listener.AddAuthorizedHandlerWithExtraAdapters(
				c.Handler,
				path,
				c.RequiredRoles,
				c.RequiredClaims,
				c.RoleComparisonOperation,
				c.ClaimComparisonOperation,
				c.ExtraAdapters,
				string(c.Method))
		} else {
			c.listener.AddAuthorizedHandlerWithRolesAndClaims(
				c.Handler,
				path,
				c.RequiredRoles,
				c.RequiredClaims,
				c.RoleComparisonOperation,
				c.ClaimComparisonOperation,
				string(c.Method))
		}
	}

	if c.NeedsAuthorization() {
		serveAuthorized(c.Path())
	} else {
		c.listener.AddHandler(c.Handler, c.Path(), string(c.Method))
	}

	if c.Version.Version != "" {
		needsDefaultApiController := true
		prefixPath := http_helper.JoinUrl(c.listener.Options.ApiPrefix, c.path)
		for _, controller := range c.listener.Controllers {
			if strings.EqualFold(controller.Path(), prefixPath) && controller.Method == c.Method {
				needsDefaultApiController = false
				break
			}
		}

		if needsDefaultApiController {
			if c.NeedsAuthorization() {
				serveAuthorized(prefixPath)
			} else {
				c.listener.AddHandler(c.Handler, prefixPath, string(c.Method))
			}
		}
	}

	return nil
}

func (c *Controller) GetHandler() ControllerHandler {
	return c.Handler
}

func (c *Controller) GetVersion() HttpVersion {
	return c.Version
}

func (c *Controller) Path() string {
	path := c.path
	if c.listener == nil {
		return ""
	}

	if c.Version.Version == "" {
		path = http_helper.JoinUrl(c.listener.Options.ApiPrefix, path)
	} else {
		path = http_helper.JoinUrl(c.listener.Options.ApiPrefix, c.Version.Path, path)
	}

	return path
}

func (c *Controller) NeedsAuthorization() bool {
	if len(c.RequiredRoles) > 0 || len(c.RequiredClaims) > 0 {
		return true
	}

	return false
}
