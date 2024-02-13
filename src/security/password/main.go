package password

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/config"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/Parallels/pd-api-service/errors"
	"golang.org/x/crypto/bcrypt"
)

var globalPasswordService *PasswordService

const (
	SPECIAL_CHARACTERS = "!@#$%^&*()_+.?"
)

type PasswordHashingAlgorithm string

const (
	PasswordHashingAlgorithmBCrypt PasswordHashingAlgorithm = "bcrypt"
	PasswordHashingAlgorithmSHA256 PasswordHashingAlgorithm = "sha256"
)

type PasswordService struct {
	ctx              basecontext.ApiContext
	HashingAlgorithm PasswordHashingAlgorithm
	options          *PasswordComplexityOptions
}

func Get() *PasswordService {
	if globalPasswordService == nil {
		ctx := basecontext.NewBaseContext()
		return New(ctx)
	}

	return globalPasswordService
}

func New(ctx basecontext.ApiContext) *PasswordService {
	globalPasswordService = &PasswordService{
		ctx:              ctx,
		HashingAlgorithm: PasswordHashingAlgorithmBCrypt,
		options:          NewPasswordComplexityOptions(ctx),
	}

	err := globalPasswordService.processEnvironmentVariables()
	if err != nil {
		ctx.LogErrorf("Error processing environment variables for password complexity options: %s", err.Error())
	}

	return globalPasswordService
}

func (s *PasswordService) GetOptions() *PasswordComplexityOptions {
	return s.options
}

func (s *PasswordService) SetOptions(options *PasswordComplexityOptions) {
	s.options = options
}

func (s *PasswordService) Hash(password string, salt string) (string, error) {
	switch s.HashingAlgorithm {
	case PasswordHashingAlgorithmBCrypt:
		return s.hashBCrypt(password, salt)
	case PasswordHashingAlgorithmSHA256:
		return s.hashSHA256(password, salt)
	default:
		s.ctx.LogErrorf("Unknown password hashing algorithm: %s", s.HashingAlgorithm)
		return "", errors.Newf("Unknown password hashing algorithm: %s", s.HashingAlgorithm)
	}
}

func (s *PasswordService) Compare(password string, salt string, hashedPwd string) error {
	switch s.HashingAlgorithm {
	case PasswordHashingAlgorithmBCrypt:
		return s.compareBCrypt(password, salt, hashedPwd)
	case PasswordHashingAlgorithmSHA256:
		return s.sha256Compare(password, salt, hashedPwd)
	default:
		s.ctx.LogErrorf("Unknown password hashing algorithm: %s", s.HashingAlgorithm)
		return errors.Newf("Unknown password hashing algorithm: %s", s.HashingAlgorithm)
	}
}

func (s *PasswordService) CheckPasswordComplexity(password string) (bool, *errors.Diagnostics) {
	diagnostics := errors.NewDiagnostics()

	if len(password) < s.options.MinLength() {
		diagnostics.AddError(errors.Newf("Password must be at least %d characters long", s.options.MinLength()))
	}
	if len(password) > s.options.MaxLength() {
		diagnostics.AddError(errors.Newf("Password must be no more than %d characters long", s.options.MaxLength()))
	}
	if s.options.RequireLowercase() {
		if !strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz") {
			diagnostics.AddError(errors.Newf("Password must contain at least one lowercase letter"))
		}
	}
	if s.options.RequireUppercase() {
		if !strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			diagnostics.AddError(errors.Newf("Password must contain at least one uppercase letter"))
		}
	}
	if s.options.RequireNumbers() {
		if !strings.ContainsAny(password, "0123456789") {
			diagnostics.AddError(errors.Newf("Password must contain at least one number"))
		}
	}
	if s.options.RequireSpecialCharacters() {
		if !strings.ContainsAny(password, SPECIAL_CHARACTERS) {
			diagnostics.AddError(errors.Newf("Password must contain at least one special character"))
		}
	}

	return !diagnostics.HasErrors(), diagnostics
}

func (s *PasswordService) hashSHA256(password string, salt string) (string, error) {
	saltedPwd, err := s.saltPassword(password, salt)
	if err != nil {
		return "", err
	}

	hashedPassword := sha256.Sum256([]byte(saltedPwd))
	return hex.EncodeToString(hashedPassword[:]), nil
}

func (s PasswordService) sha256Compare(password string, salt string, hashedPwd string) error {
	saltedPwd, err := s.saltPassword(password, salt)
	if err != nil {
		return err
	}

	hashedPassword := sha256.Sum256([]byte(saltedPwd))
	hashedPasswordString := hex.EncodeToString(hashedPassword[:])
	if hashedPasswordString != hashedPwd {
		return errors.New("passwords do not match")
	}

	return nil
}

func (s *PasswordService) hashBCrypt(password string, salt string) (string, error) {
	cost := bcrypt.DefaultCost
	saltedPwd, err := s.saltPassword(password, salt)
	if err != nil {
		return "", err
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(saltedPwd), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *PasswordService) compareBCrypt(password string, salt string, hashedPwd string) error {
	saltedPwd, err := s.saltPassword(password, salt)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPwd), saltedPwd)
	if err != nil {
		return err
	}
	return nil
}

func (s *PasswordService) saltPassword(password string, salt string) ([]byte, error) {
	// saltString := GenerateSalt(salt, cost)
	inputBytes := []byte(password)
	saltBytes := []byte(salt)
	if len(inputBytes) > 40 {
		return []byte{}, errors.New("password cannot be longer than 40 characters")
	}
	if len(saltBytes) > 32 {
		saltBytes = saltBytes[:32]
	}

	if !s.options.SaltPassword() {
		return inputBytes, nil
	}

	saltedPwd := []byte(password + string(saltBytes))

	return saltedPwd, nil
}

func (s *PasswordService) processEnvironmentVariables() error {
	cfg := config.Get()
	if cfg.GetKey(constants.SECURITY_PASSWORD_MIN_PASSWORD_LENGTH_ENV_VAR) != "" {
		minPasswordLength, err := strconv.Atoi(cfg.GetKey(constants.SECURITY_PASSWORD_MIN_PASSWORD_LENGTH_ENV_VAR))
		if err != nil {
			return err
		}
		s.options.WithMinLength(minPasswordLength)
	}
	if cfg.GetKey(constants.SECURITY_PASSWORD_MAX_PASSWORD_LENGTH_ENV_VAR) != "" {
		maxPasswordLength, err := strconv.Atoi(cfg.GetKey(constants.SECURITY_PASSWORD_MAX_PASSWORD_LENGTH_ENV_VAR))
		if err != nil {
			return err
		}
		s.options.WithMaxLength(maxPasswordLength)
	}
	if cfg.GetKey(constants.SECURITY_PASSWORD_REQUIRE_LOWERCASE_ENV_VAR) != "" {
		requireLowercase, err := strconv.ParseBool(cfg.GetKey(constants.SECURITY_PASSWORD_REQUIRE_LOWERCASE_ENV_VAR))
		if err != nil {
			return err
		}
		s.options.WithRequireLowercase(requireLowercase)
	}
	if cfg.GetKey(constants.SECURITY_PASSWORD_REQUIRE_UPPERCASE_ENV_VAR) != "" {
		requireUppercase, err := strconv.ParseBool(cfg.GetKey(constants.SECURITY_PASSWORD_REQUIRE_UPPERCASE_ENV_VAR))
		if err != nil {
			return err
		}
		s.options.WithRequireUppercase(requireUppercase)
	}
	if cfg.GetKey(constants.SECURITY_PASSWORD_REQUIRE_NUMBER_ENV_VAR) != "" {
		requireNumber, err := strconv.ParseBool(cfg.GetKey(constants.SECURITY_PASSWORD_REQUIRE_NUMBER_ENV_VAR))
		if err != nil {
			return err
		}
		s.options.WithRequireNumbers(requireNumber)
	}
	if cfg.GetKey(constants.SECURITY_PASSWORD_REQUIRE_SPECIAL_CHAR_ENV_VAR) != "" {
		requireSpecialChar, err := strconv.ParseBool(cfg.GetKey(constants.SECURITY_PASSWORD_REQUIRE_SPECIAL_CHAR_ENV_VAR))
		if err != nil {
			return err
		}
		s.options.WithRequireSpecialCharacters(requireSpecialChar)
	}
	if cfg.GetKey(constants.SECURITY_PASSWORD_SALT_PASSWORD_ENV_VAR) != "" {
		saltPassword, err := strconv.ParseBool(cfg.GetKey(constants.SECURITY_PASSWORD_SALT_PASSWORD_ENV_VAR))
		if err != nil {
			return err
		}
		s.options.WithSaltPassword(saltPassword)
	}

	return nil
}
