package helpers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"strconv"
	"strings"

	"github.com/Parallels/pd-api-service/errors"
	"golang.org/x/crypto/bcrypt"
)

func GenerateId() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func Sha256Hash(input string) (string, error) {
	hashedPassword := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hashedPassword[:]), nil
}

func BcryptHash(input string, salt string) (string, error) {
	cost := bcrypt.DefaultCost
	// saltString := GenerateSalt(salt, cost)
	inputBytes := []byte(input)
	saltBytes := []byte(salt)
	if len(inputBytes) > 40 {
		return "", errors.New("password cannot be longer than 42 characters")
	}
	if len(saltBytes) > 32 {
		saltBytes = saltBytes[:32]
	}

	saltedPwd := []byte(input + string(saltBytes))

	bytes, err := bcrypt.GenerateFromPassword([]byte(saltedPwd), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func BcryptCompare(input string, salt string, hashedPwd string) error {
	saltBytes := []byte(salt)
	if len(saltBytes) > 32 {
		saltBytes = saltBytes[:32]
	}

	saltedPwd := []byte(input + string(saltBytes))

	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), saltedPwd)
	if err != nil {
		return err
	}
	return nil
}

func ConvertByteToGigabyte(bytes float64) float64 {
	gb := float64(bytes) / 1024 / 1024 / 1024
	return math.Round(gb*100) / 100
}

func ConvertByteToMegabyte(bytes float64) float64 {
	mb := float64(bytes) / 1024 / 1024
	return math.Round(mb*100) / 100
}

func GetSizeByteFromString(s string) (float64, error) {
	s = strings.ToLower(s)
	if strings.Contains(s, "gb") || strings.Contains(s, "gi") {
		s = strings.ReplaceAll(s, "gb", "")
		s = strings.ReplaceAll(s, "gi", "")
		s = strings.TrimSpace(s)
		size, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return -1, err
		}
		return size * 1024 * 1024 * 1024, nil
	}
	if strings.Contains(s, "mb") || strings.Contains(s, "mi") {
		s = strings.ReplaceAll(s, "mb", "")
		s = strings.ReplaceAll(s, "mi", "")
		s = strings.TrimSpace(s)
		size, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return -1, err
		}
		return size * 1024 * 1024, nil
	}
	if strings.Contains(s, "kb") || strings.Contains(s, "ki") {
		s = strings.ReplaceAll(s, "kb", "")
		s = strings.ReplaceAll(s, "ki", "")
		s = strings.TrimSpace(s)
		size, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return -1, err
		}
		return size * 1024, nil
	}

	return -1, errors.New("invalid size")
}

func Obfuscate(input string) string {
	if len(input) <= 4 {
		return input
	}

	return input[0:2] + "****" + input[len(input)-2:]
}

func ContainsIllegalChars(s string) bool {
	illegalChars := []string{" ", ",", ":", ";", "(", ")", "[", "]", "{", "}", "'", "\"", "/", "\\", "|", "<", ">", "=", "+", "*", "&", "^", "%", "$", "#", "@", "!", "`", "~", "?"}
	for _, c := range illegalChars {
		if strings.Contains(s, c) {
			return true
		}
	}

	return false
}

func NormalizeString(s string) string {
	replaceChars := []string{" ", ",", ":", ";", "(", ")", "[", "]", "{", "}", "'", "\"", "/", "\\", "|", "<", ">", "=", "+", "*", "&", "^", "%", "$", "#", "@", "!", "`", "~", "?"}
	replaceWith := "_"
	for _, c := range replaceChars {
		s = strings.ReplaceAll(s, c, replaceWith)
	}

	return strings.ToLower(strings.TrimSpace(s))
}

func NormalizeStringUpper(s string) string {
	return strings.ToUpper(NormalizeString(s))
}

func CleanOutputString(s string) string {
	replaceChars := []string{"\n", "\r"}
	replaceWith := ""
	for _, c := range replaceChars {
		s = strings.ReplaceAll(s, c, replaceWith)
	}

	return strings.TrimSpace(s)
}
