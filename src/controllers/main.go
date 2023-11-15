package controllers

import (
	"fmt"

	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/restapi"
)

func RegisterV1() error {
	fmt.Println("Registering V1")
	apiKeysTestController := restapi.NewController()
	apiKeysTestController.WithMethod(restapi.GET)
	apiKeysTestController.WithVersion("v1")
	apiKeysTestController.WithPath("/test/api_keys")
	apiKeysTestController.WithRequiredRole(constants.SUPER_USER_ROLE)
	apiKeysTestController.WithHandler(GetApiKeysController()).Register()

	return nil
}
