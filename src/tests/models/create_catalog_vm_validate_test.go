package models_test

import (
	"testing"

	"github.com/Parallels/prl-devops-service/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateCatalogVirtualMachineRequest_Validate covers the updated
// validation logic that no longer requires connection or catalog_manager_id.

func TestValidate_MissingCatalogId(t *testing.T) {
	req := &models.CreateCatalogVirtualMachineRequest{
		MachineName: "test-vm",
		Connection:  "http://catalog",
	}
	err := req.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing catalog id")
}

func TestValidate_MissingMachineName(t *testing.T) {
	req := &models.CreateCatalogVirtualMachineRequest{
		CatalogId:  "my-catalog",
		Connection: "http://catalog",
	}
	err := req.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing machine name")
}

func TestValidate_BothConnectionAndManagerId(t *testing.T) {
	req := &models.CreateCatalogVirtualMachineRequest{
		CatalogId:        "my-catalog",
		MachineName:      "test-vm",
		Connection:       "http://catalog",
		CatalogManagerId: "mgr-id",
	}
	err := req.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot both be provided")
}

func TestValidate_ConnectionOnly(t *testing.T) {
	req := &models.CreateCatalogVirtualMachineRequest{
		CatalogId:   "my-catalog",
		MachineName: "test-vm",
		Connection:  "http://catalog",
	}
	err := req.Validate()
	require.NoError(t, err)
}

func TestValidate_CatalogManagerIdOnly(t *testing.T) {
	req := &models.CreateCatalogVirtualMachineRequest{
		CatalogId:        "my-catalog",
		MachineName:      "test-vm",
		CatalogManagerId: "mgr-id",
	}
	err := req.Validate()
	require.NoError(t, err)
}

// TestValidate_NeitherConnectionNorManagerId verifies that omitting both
// connection and catalog_manager_id is now valid (local catalog fallback).
func TestValidate_NeitherConnectionNorManagerId(t *testing.T) {
	req := &models.CreateCatalogVirtualMachineRequest{
		CatalogId:   "my-catalog",
		MachineName: "test-vm",
	}
	err := req.Validate()
	require.NoError(t, err, "omitting connection and catalog_manager_id must be valid for local catalog fallback")
}

func TestValidate_SetsDefaultVersion(t *testing.T) {
	req := &models.CreateCatalogVirtualMachineRequest{
		CatalogId:   "my-catalog",
		MachineName: "test-vm",
	}
	require.NoError(t, req.Validate())
	assert.NotEmpty(t, req.Version, "version should be defaulted to latest tag")
}
