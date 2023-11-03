package controllers

import (
	"Parallels/pd-api-service/mappers"
	"Parallels/pd-api-service/models"
	"Parallels/pd-api-service/restapi"
	"encoding/json"
	"net/http"

	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func GetUsersController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		users, err := dbService.GetUsers(ctx, GetFilterHeader(r))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		if len(users) == 0 {
			w.WriteHeader(http.StatusOK)
			response := make([]models.ApiUser, 0)
			json.NewEncoder(w).Encode(response)
			ctx.LogInfo("Users returned: %v", len(response))
			return
		}

		result := mappers.DtoUsersToApiResponse(users)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Users returned: %v", len(result))
	}
}

func GetUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		user, err := dbService.GetUser(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoUserToApiResponse(*user)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("User returned: %v", response.ID)
	}
}

func CreateUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.UserCreateRequest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		dtoUser, err := dbService.CreateUser(ctx, mappers.ApiUserCreateRequestToDto(request))
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		response := mappers.DtoUserToApiResponse(*dtoUser)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		ctx.LogInfo("User created: %v", response.ID)
	}
}

func DeleteUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		err = dbService.DeleteUser(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("User deleted: %v", id)
	}
}

func UpdateUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.UserCreateRequest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		dtoUser := mappers.ApiUserCreateRequestToDto(request)
		dtoUser.ID = id
		err = dbService.UpdateUser(ctx, dtoUser)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("User updated: %v", id)
	}
}

func GetUserRolesController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		dtoUser, err := dbService.GetUser(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		roles := dtoUser.Roles
		result := mappers.DtoRolesToApi(roles)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Roles returned: %v", len(result))
	}
}

func AddRoleToUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.RoleRequest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		if err := dbService.AddRoleToUser(ctx, id, request.Name); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		dbService.Disconnect(ctx)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(request)
		ctx.LogInfo("Role added to user: %v", id)
	}
}

func RemoveRoleFromUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		roleId := vars["role_id"]

		if err = dbService.RemoveRoleFromUser(ctx, id, roleId); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Role removed from user: %v", id)
	}
}

func GetUserClaimsController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		dtoUser, err := dbService.GetUser(ctx, id)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		claims := dtoUser.Claims
		result := mappers.DtoClaimsToApi(claims)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
		ctx.LogInfo("Claims returned: %v", len(result))
	}
}

func AddClaimToUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		var request models.ClaimRequest
		http_helper.MapRequestBody(r, &request)
		if err := request.Validate(); err != nil {
			ReturnApiError(ctx, w, models.ApiErrorResponse{
				Message: "Invalid request body: " + err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]

		if err := dbService.AddClaimToUser(ctx, id, request.Name); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		dbService.Disconnect(ctx)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(request)
		ctx.LogInfo("Claim added to user: %v", id)
	}
}

func RemoveClaimFromUserController() restapi.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := GetBaseContext(r)
		dbService, err := GetDatabaseService(ctx)
		if err != nil {
			ReturnApiError(ctx, w, models.NewFromErrorWithCode(err, http.StatusInternalServerError))
			return
		}

		defer dbService.Disconnect(ctx)

		vars := mux.Vars(r)
		id := vars["id"]
		claimId := vars["claim_id"]

		if err = dbService.RemoveClaimFromUser(ctx, id, claimId); err != nil {
			ReturnApiError(ctx, w, models.NewFromError(err))
			return
		}

		w.WriteHeader(http.StatusAccepted)
		ctx.LogInfo("Claim removed from user: %v", id)
	}
}
