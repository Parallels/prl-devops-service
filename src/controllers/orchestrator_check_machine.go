package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/Parallels/prl-devops-service/models"
	"github.com/Parallels/prl-devops-service/orchestrator"
	"github.com/Parallels/prl-devops-service/restapi"

	"github.com/cjlapao/common-go/helper/http_helper"
)

func registerOrchestratorMachineCheckHandlers(ctx basecontext.ApiContext, version string) {
	ctx.LogInfof("Registering version %s Orchestrator machine requirements check handlers", version)
	restapi.NewController().
		WithMethod(restapi.POST).
		WithVersion(version).
		WithPath("/orchestrator/machines/check").
		WithRequiredClaim(constants.CREATE_VM_CLAIM).
		WithHandler(CheckOrchestratorMachineRequirementsHandler()).
		Register()
}

// CheckOrchestratorMachineRequirementsHandler runs the scheduling decision factor
// ahead of VM creation. The requirements are taken either from a catalog manifest
// (catalog_id, resolved through the catalog manager exactly like POST machines) or
// from the request body (requirements, including the OS to know whether this is a
// macOS VM). When any host has enough resources the response is slim; otherwise it
// details, per host, exactly what is missing (disk space, CPU, memory, per-host VM
// count limit, Apple VM capacity, ...).
//
// @Summary		Check host requirements before creating a virtual machine
// @Description	Evaluates the VM scheduling decision factor against the registered hosts. Requirements are resolved from a catalog manifest (catalog_id + version + architecture, served by a catalog manager) or provided directly in the body. The response is slim when resources are sufficient and carries the full per-host missing-resource detail when no host qualifies.
// @Tags			Orchestrator
// @Accept			json
// @Produce		json
// @Param			request	body		models.CheckMachineRequirementsRequest	true	"Requirements check request"
// @Success		200		{object}	models.CheckMachineRequirementsResponse
// @Failure		400		{object}	models.ApiErrorResponse
// @Failure		401		{object}	models.OAuthErrorResponse
// @Failure		404		{object}	models.ApiErrorResponse
// @Security		ApiKeyAuth
// @Security		BearerAuth
// @Router			/v1/orchestrator/machines/check [post]
func CheckOrchestratorMachineRequirementsHandler() restapi.ControllerHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		ctx := GetBaseContext(r)
		defer Recover(ctx, r, w)

		var request models.CheckMachineRequirementsRequest
		if err := http_helper.MapRequestBody(r, &request); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		// Catalog-driven checks resolve the connection through the catalog manager,
		// sharing VM creation's authorization and local-catalog semantics. No
		// temporary credentials are needed: manifest resolution happens in-process.
		if request.CatalogId != "" {
			catalogRequest := &models.CreateCatalogVirtualMachineRequest{
				CatalogId:        request.CatalogId,
				Version:          request.Version,
				Architecture:     request.Architecture,
				Connection:       request.Connection,
				CatalogManagerId: request.CatalogManagerId,
				Owner:            request.Owner,
			}
			catalogConnection, err := resolveCatalogMachineConnection(ctx, catalogRequest)
			if err != nil {
				ReturnApiError(ctx, w, models.NewFromError(err))
				return
			}
			request.Connection = catalogConnection
			request.CatalogManagerId = ""
		}

		orchestratorSvc := orchestrator.NewOrchestratorService(ctx)
		if orchestratorSvc == nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Orchestrator is not available",
				Code:    http.StatusInternalServerError,
			})
			return
		}

		response, apiError := orchestratorSvc.CheckMachineRequirements(ctx, request)
		if apiError != nil {
			ReturnApiError(ctx, w, *apiError)
			return
		}

		if response.Passed {
			ctx.LogInfof("Machine requirements check passed")
		} else {
			ctx.LogInfof("Machine requirements check failed: %s", response.Message)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}
}
