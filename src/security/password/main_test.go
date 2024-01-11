package password

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/Parallels/pd-api-service/constants"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestNoHashingAlgorithm(t *testing.T) {
	input := "password"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = "test"

	_, err := svc.Hash(input, salt)
	if err != nil {
		assert.EqualError(t, err, "error: Unknown password hashing algorithm: test")
	}
}

func TestGetWithoutNew(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	globalPasswordService = nil

	svc := Get()
	svc.SetOptions(&PasswordComplexityOptions{
		ctx:                      ctx,
		minLength:                20,
		maxLength:                40,
		requireLowercase:         true,
		requireUppercase:         true,
		requireNumbers:           true,
		requireSpecialCharacters: true,
		saltPassword:             false,
	})

	testSvc := Get()

	assert.Equal(t, 20, svc.GetOptions().MinLength())
	assert.Equal(t, 40, svc.GetOptions().MaxLength())
	assert.Equal(t, true, svc.GetOptions().RequireLowercase())
	assert.Equal(t, true, svc.GetOptions().RequireUppercase())
	assert.Equal(t, true, svc.GetOptions().RequireNumbers())
	assert.Equal(t, true, svc.GetOptions().RequireSpecialCharacters())

	assert.Equal(t, testSvc.GetOptions().MinLength(), svc.GetOptions().MinLength())
	assert.Equal(t, testSvc.GetOptions().MaxLength(), svc.GetOptions().MaxLength())
	assert.Equal(t, testSvc.GetOptions().RequireLowercase(), svc.GetOptions().RequireLowercase())
	assert.Equal(t, testSvc.GetOptions().RequireUppercase(), svc.GetOptions().RequireUppercase())
	assert.Equal(t, testSvc.GetOptions().RequireNumbers(), svc.GetOptions().RequireNumbers())
	assert.Equal(t, testSvc.GetOptions().RequireSpecialCharacters(), svc.GetOptions().RequireSpecialCharacters())
	assert.Equal(t, testSvc.GetOptions().SaltPassword(), svc.GetOptions().SaltPassword())
}

func TestSHA256hash(t *testing.T) {
	input := "password"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = PasswordHashingAlgorithmSHA256

	hashedPwd, err := svc.Hash(input, salt)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}

	CompareHashedPwd := sha256.Sum256([]byte(input + salt))
	hashedPwdStr := hex.EncodeToString(CompareHashedPwd[:])

	assert.Equal(t, hashedPwd, hashedPwdStr)
}

func TestBcryptHash(t *testing.T) {
	input := "password"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = PasswordHashingAlgorithmBCrypt

	hashedPwd, err := svc.Hash(input, salt)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(input+salt))
	if err != nil {
		t.Errorf("Hashed password does not match input: %v", err)
	}
}

func TestBcryptHashWithNoSalt(t *testing.T) {
	input := "password"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = PasswordHashingAlgorithmBCrypt
	svc.GetOptions().WithSaltPassword(false)

	hashedPwd, err := svc.Hash(input, "")
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(input))
	if err != nil {
		t.Errorf("Hashed password does not match input: %v", err)
	}
}

func TestSaltPasswordBiggerThan40Characters(t *testing.T) {
	input := "password"
	salt := "somesaltthatisbiggerthan40characters"
	smallerSalt := "somesaltthatisbiggerthan40charac"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = PasswordHashingAlgorithmBCrypt

	hashedPwd, err := svc.Hash(input, salt)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(input+smallerSalt))
	if err != nil {
		t.Errorf("Hashed password does not match input: %v", err)
	}
}

func TestPasswordLongerThan40Characters(t *testing.T) {
	input := "passwordpasswordpasswordpasswordpasswordpassword"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = PasswordHashingAlgorithmBCrypt

	_, err := svc.Hash(input, salt)
	if err != nil {
		assert.EqualError(t, err, "error: password cannot be longer than 40 characters")
	}
}

func TestSetOptionsWithValuesGreaterThan40(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = PasswordHashingAlgorithmBCrypt

	svc.GetOptions().WithMinLength(4)
	svc.GetOptions().WithMaxLength(50)

	assert.Equal(t, 8, svc.GetOptions().MinLength())
	assert.Equal(t, 40, svc.GetOptions().MaxLength())
}

func TestCheckPasswordComplexity(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.GetOptions().WithMinLength(8)
	svc.GetOptions().WithMaxLength(40)
	svc.GetOptions().WithRequireLowercase(true)
	svc.GetOptions().WithRequireUppercase(true)
	svc.GetOptions().WithRequireNumbers(true)
	svc.GetOptions().WithRequireSpecialCharacters(true)

	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{
			name:     "Valid password",
			password: "Password123!",
			expected: true,
		},
		{
			name:     "Password too short",
			password: "pass",
			expected: false,
		},
		{
			name:     "Password too long",
			password: "passwordpasswordpasswordpasswordpasswordpasswordpasswordpassword",
			expected: false,
		},
		{
			name:     "Missing lowercase letter",
			password: "PASSWORD123!",
			expected: false,
		},
		{
			name:     "Missing uppercase letter",
			password: "password123!",
			expected: false,
		},
		{
			name:     "Missing number",
			password: "Password!",
			expected: false,
		},
		{
			name:     "Missing special character",
			password: "Password123",
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			valid, diagnostics := svc.CheckPasswordComplexity(test.password)
			assert.Equal(t, test.expected, valid)
			if test.expected {
				assert.Empty(t, diagnostics.Errors())
			} else {
				assert.NotEmpty(t, diagnostics.Errors())
			}
		})
	}
}

func TestCompareBcrypt(t *testing.T) {
	input := "password"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = PasswordHashingAlgorithmBCrypt

	hashedPwd, err := svc.Hash(input, salt)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}

	err = svc.Compare(input, salt, hashedPwd)
	if err != nil {
		t.Errorf("Hashed password does not match input: %v", err)
	}
}

func TestCompareSHA256(t *testing.T) {
	input := "password"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = PasswordHashingAlgorithmSHA256

	hashedPwd, err := svc.Hash(input, salt)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}

	err = svc.Compare(input, salt, hashedPwd)
	if err != nil {
		t.Errorf("Hashed password does not match input: %v", err)
	}
}

func TestCompareSHA256WithWrongPassword(t *testing.T) {
	input := "password"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = PasswordHashingAlgorithmSHA256

	hashedPwd, err := svc.Hash(input, salt)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}

	err = svc.Compare("wrongpassword", salt, hashedPwd)
	assert.EqualError(t, err, "error: passwords do not match")
}

func TestCompareWithNoHash(t *testing.T) {
	input := "password"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = "test"

	err := svc.Compare(input, salt, "")
	assert.EqualError(t, err, "error: Unknown password hashing algorithm: test")
}

func TestHashWithBcryptSaltError(t *testing.T) {
	input := "passwordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpassword"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)

	_, err := svc.Hash(input, salt)
	assert.EqualError(t, err, "error: password cannot be longer than 40 characters")
}

func TestHashWithSHA256SaltError(t *testing.T) {
	input := "passwordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpassword"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = PasswordHashingAlgorithmSHA256

	_, err := svc.Hash(input, salt)
	assert.EqualError(t, err, "error: password cannot be longer than 40 characters")
}

func TestCompareWithBcryptSaltError(t *testing.T) {
	input := "password"
	LongInput := "passwordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpassword"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)

	_, err := svc.Hash(input, salt)
	assert.Nil(t, err)
	err = svc.Compare(LongInput, salt, "")
	assert.EqualError(t, err, "error: password cannot be longer than 40 characters")
}

func TestCompareWithBcryptError(t *testing.T) {
	input := "password"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)

	_, err := svc.Hash(input, salt)
	assert.Nil(t, err)
	err = svc.Compare("", salt, "")
	assert.EqualError(t, err, "crypto/bcrypt: hashedSecret too short to be a bcrypted password")
}

func TestCompareWithSHA256SaltError(t *testing.T) {
	input := "password"
	LongInput := "passwordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpasswordpassword"
	salt := "somesalt"
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.HashingAlgorithm = PasswordHashingAlgorithmSHA256

	_, err := svc.Hash(input, salt)
	assert.Nil(t, err)
	err = svc.Compare(LongInput, salt, "")
	assert.EqualError(t, err, "error: password cannot be longer than 40 characters")
}
func TestPasswordService_processEnvironmentVariables(t *testing.T) {
	svc := New(nil)

	t.Run("SetMinPasswordLength", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.SECURITY_PASSWORD_MIN_PASSWORD_LENGTH_ENV_VAR, "8")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.NoError(t, err)

		assert.Equal(t, 8, svc.GetOptions().MinLength())
	})

	t.Run("SetMaxPasswordLength", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.SECURITY_PASSWORD_MAX_PASSWORD_LENGTH_ENV_VAR, "40")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.NoError(t, err)

		assert.Equal(t, 40, svc.GetOptions().MaxLength())
	})

	t.Run("SetRequireLowercase", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.SECURITY_PASSWORD_REQUIRE_LOWERCASE_ENV_VAR, "true")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.NoError(t, err)

		assert.Equal(t, true, svc.GetOptions().RequireLowercase())
	})

	t.Run("SetRequireUppercase", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.SECURITY_PASSWORD_REQUIRE_UPPERCASE_ENV_VAR, "true")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.NoError(t, err)

		assert.Equal(t, true, svc.GetOptions().RequireUppercase())
	})

	t.Run("SetRequireNumber", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.SECURITY_PASSWORD_REQUIRE_NUMBER_ENV_VAR, "true")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.NoError(t, err)

		assert.Equal(t, true, svc.GetOptions().RequireNumbers())
	})

	t.Run("SetRequireSpecialChar", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.SECURITY_PASSWORD_REQUIRE_SPECIAL_CHAR_ENV_VAR, "true")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.NoError(t, err)

		assert.Equal(t, true, svc.GetOptions().RequireSpecialCharacters())
	})

	t.Run("SetSaltPassword", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.SECURITY_PASSWORD_SALT_PASSWORD_ENV_VAR, "true")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.NoError(t, err)

		assert.Equal(t, true, svc.GetOptions().SaltPassword())
	})
}

func TestPasswordService_processEnvironmentVariablesError(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	t.Run("SetMinPasswordLengthError", func(t *testing.T) {
		os.Clearenv()
		svc := New(ctx)

		err := os.Setenv(constants.SECURITY_PASSWORD_MIN_PASSWORD_LENGTH_ENV_VAR, "A")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.Errorf(t, err, "strconv.Atoi: parsing \"A\": invalid syntax")

		assert.Equal(t, 12, svc.GetOptions().MinLength())
	})

	t.Run("SetMaxPasswordLength", func(t *testing.T) {
		os.Clearenv()
		svc := New(ctx)

		err := os.Setenv(constants.SECURITY_PASSWORD_MAX_PASSWORD_LENGTH_ENV_VAR, "A")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.Errorf(t, err, "strconv.Atoi: parsing \"A\": invalid syntax")

		assert.Equal(t, 40, svc.GetOptions().MaxLength())
	})

	t.Run("SetRequireLowercase", func(t *testing.T) {
		os.Clearenv()
		svc := New(ctx)

		err := os.Setenv(constants.SECURITY_PASSWORD_REQUIRE_LOWERCASE_ENV_VAR, "A")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.Errorf(t, err, "strconv.Atoi: parsing \"A\": invalid syntax")

		assert.Equal(t, true, svc.GetOptions().RequireLowercase())
	})

	t.Run("SetRequireUppercase", func(t *testing.T) {
		os.Clearenv()
		svc := New(ctx)

		err := os.Setenv(constants.SECURITY_PASSWORD_REQUIRE_UPPERCASE_ENV_VAR, "A")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.Errorf(t, err, "strconv.Atoi: parsing \"A\": invalid syntax")

		assert.Equal(t, true, svc.GetOptions().RequireUppercase())
	})

	t.Run("SetRequireNumber", func(t *testing.T) {
		os.Clearenv()
		svc := New(ctx)

		err := os.Setenv(constants.SECURITY_PASSWORD_REQUIRE_NUMBER_ENV_VAR, "A")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.Errorf(t, err, "strconv.Atoi: parsing \"A\": invalid syntax")

		assert.Equal(t, true, svc.GetOptions().RequireNumbers())
	})

	t.Run("SetRequireSpecialChar", func(t *testing.T) {
		os.Clearenv()
		svc := New(ctx)

		err := os.Setenv(constants.SECURITY_PASSWORD_REQUIRE_SPECIAL_CHAR_ENV_VAR, "A")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.Errorf(t, err, "strconv.Atoi: parsing \"A\": invalid syntax")

		assert.Equal(t, true, svc.GetOptions().RequireSpecialCharacters())
	})

	t.Run("SetSaltPassword", func(t *testing.T) {
		os.Clearenv()
		svc := New(ctx)

		err := os.Setenv(constants.SECURITY_PASSWORD_SALT_PASSWORD_ENV_VAR, "A")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.Errorf(t, err, "strconv.Atoi: parsing \"A\": invalid syntax")

		assert.Equal(t, true, svc.GetOptions().SaltPassword())
	})
}
