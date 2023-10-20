package helpers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateId() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func Sha256Hash(input string) string {
	hashedPassword := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hashedPassword[:])
}
