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
	listener       *HttpListener
	path           string
	Handler        ControllerHandler
	Version        HttpVersion
	Method         HttpControllerMethod
	RequiredRoles  []string
	RequiredClaims []string
}

// NewController creates a new instance of the Controller struct with default values.
// The returned controller has the GET method, uses the global HTTP listener, and has no required roles or claims.
func NewController() *Controller {
	controller := &Controller{
		Method:         GET,
		listener:       globalHttpListener,
		RequiredRoles:  make([]string, 0),
		RequiredClaims: make([]string, 0),
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

func (c *Controller) WithHandler(handler ControllerHandler) *Controller {
	c.Handler = handler
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

	if c.NeedsAuthorization() {
		c.listener.AddAuthorizedHandlerWithRolesAndClaims(c.Handler, c.Path(), c.RequiredRoles, c.RequiredClaims, string(c.Method))
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
				c.listener.AddAuthorizedHandlerWithRolesAndClaims(c.Handler, prefixPath, c.RequiredRoles, c.RequiredClaims, string(c.Method))
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
