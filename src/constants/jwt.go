package constants

const (
	JWT_PRIVATE_KEY_ENV_VAR    = "JWT_PRIVATE_KEY"
	JWT_HMACS_SECRET_ENV_VAR   = "JWT_HMACS_SECRET" // #nosec G101 This is not a hardcoded password, it is just the variable name we use to store the secret
	JWT_DURATION_ENV_VAR       = "JWT_DURATION"
	JWT_SIGN_ALGORITHM_ENV_VAR = "JWT_SIGN_ALGORITHM"
)
