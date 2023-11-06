package helpers

import (
	"testing"
)

func TestCreateDirIfNotExist(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test creating a directory that doesn't exist
	err := CreateDirIfNotExist(tempDir + "/newdir")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Test creating a directory that already exists
	err = CreateDirIfNotExist(tempDir + "/newdir")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// // Test creating a directory with an invalid path
	// err = CreateDirIfNotExist(tempDir + "/:newdir/invalid/path")
	// if err == nil {
	// 	t.Errorf("Expected an error, but got none")
	// }
	// if !os.IsNotExist(err) {
	// 	t.Errorf("Expected a 'not exist' error, but got %v", err)
	// }
}
