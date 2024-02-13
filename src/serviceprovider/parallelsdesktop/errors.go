package parallelsdesktop

import "github.com/Parallels/prl-devops-service/errors"

var ErrVirtualMachineNotFound = errors.NewWithCode("Virtual machine not found", 404)
