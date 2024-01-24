package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Parallels/pd-api-service/errors"
)

func GenerateId() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return hex.EncodeToString(bytes)
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

func SanitizeArchivePath(d, t string) (v string, err error) {
	v = filepath.Join(d, t)
	if strings.HasPrefix(v, filepath.Clean(d)) {
		return v, nil
	}

	return "", fmt.Errorf("%s: %s", "content filepath is tainted", t)
}
