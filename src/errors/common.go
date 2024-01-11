package errors

func ErrNotFound() error {
	return NewWithCode("not found", 404)
}

func ErrValueEmpty() error {
	return NewWithCode("value cannot be empty", 400)
}

func ErrMissingId() error {
	return NewWithCode("missing id", 404)
}

func ErrMissingPath() error {
	return NewWithCode("missing path", 404)
}

func ErrNoSystemUserFound() error {
	return NewWithCode("no system user found", 404)
}

func ErrInvalidFilter() error {
	return NewWithCode("invalid filter", 400)
}

func ErrInvalidFilterProperty() error {
	return NewWithCode("invalid filter property", 400)
}

func ErrNoVirtualMachineFound(id string) error {
	return NewWithCodef(404, "no virtual machine found with %s", id)
}

func ErrNoVirtualMachinesFound() error {
	return NewWithCode("no virtual machines found", 404)
}

func ErrConfigOperationEmpty() error {
	return NewWithCode("config operation cannot be empty", 400)
}

func ErrConfigGroupEmpty() error {
	return NewWithCode("config group cannot be empty", 400)
}

func ErrConfigOperationNotSupported(group, operation string) error {
	return NewWithCodef(400, "operation %s not supported on group %s", operation, group)
}

func ErrConfigOperationNoEnoughArguments(group, operation string) error {
	return NewWithCodef(400, "operation %s does not have enough arguments on group %s", operation, group)
}

func ErrConfigInvalidOperation(operation string) error {
	return NewWithCodef(400, "invalid operation %s", operation)
}

func ErrConfigInvalidBiosType(v string) error {
	return NewWithCodef(400, "Invalid BIOS type %s", v)
}

func ErrConfigMissingSharedFolderPath() error {
	return NewWithCode("missing shared folder path", 400)
}
