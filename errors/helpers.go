package errors

func GetSystemErrorCode(err error) int {
	_, ok := err.(*SystemError)
	if !ok {
		_, ok = err.(SystemError)
		if ok {
			code := err.(SystemError).Code()
			return code
		}

		return 0
	}

	code := err.(*SystemError).Code()
	return code
}
