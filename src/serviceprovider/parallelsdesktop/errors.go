package parallelsdesktop

import "github.com/Parallels/prl-devops-service/errors"

var ErrVirtualMachineNotFound = errors.NewWithCode("Virtual machine not found", 404)
var ErrVirtualMachineNotFoundInCache = errors.NewWithCode("Virtual machine not found in cache even after multiple attempts", 404)
