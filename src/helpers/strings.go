package helpers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Parallels/prl-devops-service/errors"
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
	if s == "" {
		return -1, errors.New("size string is empty")
	}

	s = strings.ToLower(s)
	s = strings.TrimSpace(s)

	// Define units with their suffixes and multipliers
	// Order matters: check multi-letter suffixes before single-letter ones
	units := []struct {
		suffixes   []string
		multiplier float64
	}{
		{[]string{"tb", "ti"}, 1024 * 1024 * 1024 * 1024},
		{[]string{"gb", "gi"}, 1024 * 1024 * 1024},
		{[]string{"mb", "mi"}, 1024 * 1024},
		{[]string{"kb", "ki"}, 1024},
		{[]string{"bi", "b"}, 1},
		{[]string{"t"}, 1024 * 1024 * 1024 * 1024},
		{[]string{"g"}, 1024 * 1024 * 1024},
		{[]string{"m"}, 1024 * 1024},
		{[]string{"k"}, 1024},
	}

	for _, unit := range units {
		for _, suffix := range unit.suffixes {
			if strings.HasSuffix(s, suffix) {
				numStr := strings.TrimSuffix(s, suffix)
				numStr = strings.TrimSpace(numStr)
				size, err := strconv.ParseFloat(numStr, 64)
				if err != nil {
					return -1, err
				}
				return size * unit.multiplier, nil
			}
		}
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

func ObfuscateString(value string) string {
	if value == "" {
		return value
	}

	if len(value) <= 4 {
		return value
	}

	return value[0:2] + "****" + value[len(value)-2:]
}

func ClearLine() {
	fmt.Printf("\r\033[K")
}

func StringToBool(s string) bool {
	if s == "true" ||
		s == "1" ||
		s == "yes" ||
		s == "y" ||
		s == "t" ||
		s == "on" ||
		s == "enable" ||
		s == "enabled" ||
		s == "active" {
		return true
	}

	return false
}

func ConvertCompressRatioFromString(ratio string) (int, error) {
	switch strings.ToLower(ratio) {
	case "best_speed":
		return 1, nil
	case "balanced":
		return 5, nil
	case "best_compression":
		return 9, nil
	case "no_compression":
		return 0, nil
	case "default":
		return -1, nil
	default:
		return -1, errors.New("invalid compression ratio")
	}
}

func GetCompressRatioEnvValue(ratioValue int) (string, error) {
	switch ratioValue {
	case 1:
		return "best_speed", nil
	case 5:
		return "balanced", nil
	case 9:
		return "best_compression", nil
	case 0:
		return "no_compression", nil
	case -1:
		return "default", nil
	default:
		return "", errors.New("invalid compression ratio")
	}
}
