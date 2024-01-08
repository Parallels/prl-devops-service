package helpers

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHash(t *testing.T) {
	input := "password"
	salt := "somesalt"

	hashedPwd, err := BcryptHash(input, salt)
	if err != nil {
		t.Errorf("Error hashing password: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(input+salt))
	if err != nil {
		t.Errorf("Hashed password does not match input: %v", err)
	}
}
