package restapi

import (
	"testing"
)

func TestNewController(t *testing.T) {
	controller := NewController()

	if controller.Method != GET {
		t.Errorf("Expected method %s, but got %s", GET, controller.Method)
	}

	if controller.listener != globalHttpListener {
		t.Errorf("Expected listener %v, but got %v", globalHttpListener, controller.listener)
	}

	if len(controller.RequiredRoles) != 0 {
		t.Errorf("Expected RequiredRoles to be empty, but got %v", controller.RequiredRoles)
	}

	if len(controller.RequiredClaims) != 0 {
		t.Errorf("Expected RequiredClaims to be empty, but got %v", controller.RequiredClaims)
	}
}

func TestController_WithPath(t *testing.T) {
	controller := NewController()
	path := "/test"

	// Call the WithPath method
	controller.WithPath(path)

	// Check if the path was set correctly
	if controller.path != path {
		t.Errorf("Expected path %s, but got %s", path, controller.path)
	}
}

func TestController_WithVersion(t *testing.T) {
	controller := NewController()
	controller.listener = &HttpListener{
		Versions: []HttpVersion{
			{
				Path: "/v1",
			},
		},
	}

	versionPath := "/v1"

	// Call the WithVersion method
	controller.WithVersion(versionPath)

	// Check if the version was set correctly
	if controller.Version.Version != "" {
		t.Errorf("Expected version to be set, but it was nil")
	}

	if controller.Version.Path != versionPath {
		t.Errorf("Expected version path %s, but got %s", versionPath, controller.Version.Path)
	}
}

func TestController_WithMethod(t *testing.T) {
	controller := NewController()
	method := POST

	// Call the WithMethod method
	controller.WithMethod(method)

	// Check if the method was set correctly
	if controller.Method != method {
		t.Errorf("Expected method %s, but got %s", method, controller.Method)
	}
}

func TestController_Path(t *testing.T) {
	controller := NewController()
	controller.WithPath("/test")
	controller.listener = &HttpListener{
		Options: &HttpListenerOptions{
			ApiPrefix: "/api",
		},
	}

	expectedPath := "/api/test"

	// Call the Path method
	path := controller.Path()

	// Check if the path was generated correctly
	if path != expectedPath {
		t.Errorf("Expected path %s, but got %s", expectedPath, path)
	}

	// Test with version
	controller.listener.Versions = []HttpVersion{
		{
			Version: "v1",
			Path:    "/v1",
		},
	}
	controller.WithVersion("/v1")

	expectedPath = "/api/v1/test"

	// Call the Path method
	path = controller.Path()

	// Check if the path was generated correctly
	if path != expectedPath {
		t.Errorf("Expected path %s, but got %s", expectedPath, path)
	}
}
