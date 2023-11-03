package parallelsdesktop

import "Parallels/pd-api-service/errors"

var (
	ErrVirtualMachineNotFound = errors.NewWithCode("Virtual machine not found", 404)
)
