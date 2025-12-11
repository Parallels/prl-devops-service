package eventemitter

import (
	"net/http"
	"testing"

	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringToEventTypes_ValidTypes(t *testing.T) {
	result, err := stringToEventTypes([]string{"pdfm", "system", "global"})
	require.NoError(t, err)
	expected := []constants.EventType{
		constants.EventTypePDFM,
		constants.EventTypeSystem,
		constants.EventTypeGlobal,
	}
	assert.Equal(t, expected, result)
}

func TestStringToEventTypes_EmptySlice(t *testing.T) {
	result, err := stringToEventTypes([]string{})
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestStringToEventTypes_InvalidTypes(t *testing.T) {
	result, err := stringToEventTypes([]string{"invalid", "pdfm", "fake"})
	// Should return error but still include valid types
	assert.Error(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, constants.EventTypePDFM, result[0])
}

func TestStringToEventTypes_MixedCase(t *testing.T) {
	result, err := stringToEventTypes([]string{"PDFM", "System", "GLOBAL"})
	require.NoError(t, err)
	expected := []constants.EventType{
		constants.EventTypePDFM,
		constants.EventTypeSystem,
		constants.EventTypeGlobal,
	}
	assert.Equal(t, expected, result)
}

func TestStringToEventTypes_ExtraWhitespace(t *testing.T) {
	result, err := stringToEventTypes([]string{"  pdfm  ", " system ", " global "})
	require.NoError(t, err)
	expected := []constants.EventType{
		constants.EventTypePDFM,
		constants.EventTypeSystem,
		constants.EventTypeGlobal,
	}
	assert.Equal(t, expected, result)
}

func TestStringToEventTypes_AllValidTypes(t *testing.T) {
	result, err := stringToEventTypes([]string{"global", "system", "pdfm"})
	require.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Contains(t, result, constants.EventTypeGlobal)
	assert.Contains(t, result, constants.EventTypeSystem)
	assert.Contains(t, result, constants.EventTypePDFM)
}

func TestStringToEventTypes_OnlyInvalidTypes(t *testing.T) {
	result, err := stringToEventTypes([]string{"invalid", "fake", "bad"})
	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestExtractClientIP_XRealIP(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-Ip", "203.0.113.195")

	ip := extractClientIP(req)
	assert.Equal(t, "203.0.113.195", ip)
}

func TestExtractClientIP_XForwardedFor_Single(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "198.51.100.1")

	ip := extractClientIP(req)
	assert.Equal(t, "198.51.100.1", ip)
}

func TestExtractClientIP_XForwardedFor_Multiple(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "198.51.100.1, 203.0.113.195, 192.0.2.1")

	ip := extractClientIP(req)
	// Should return the first IP in the chain
	assert.Equal(t, "198.51.100.1", ip)
}

func TestExtractClientIP_RemoteAddr(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	ip := extractClientIP(req)
	assert.Equal(t, "192.168.1.100", ip)
}

func TestExtractClientIP_RemoteAddr_IPv6(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "[2001:db8::1]:12345"

	ip := extractClientIP(req)
	// LastIndex finds the colon after the IPv6 address, keeps the brackets
	assert.Equal(t, "[2001:db8::1]", ip)
}

func TestExtractClientIP_PreferXForwardedForOverXRealIP(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Real-Ip", "203.0.113.195")
	req.Header.Set("X-Forwarded-For", "198.51.100.1")
	req.RemoteAddr = "192.168.1.100:12345"

	ip := extractClientIP(req)
	// X-Forwarded-For takes precedence over X-Real-IP
	assert.Equal(t, "198.51.100.1", ip)
}

func TestExtractClientIP_PreferXForwardedFor(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "198.51.100.1")
	req.RemoteAddr = "192.168.1.100:12345"

	ip := extractClientIP(req)
	// X-Forwarded-For should take precedence over RemoteAddr
	assert.Equal(t, "198.51.100.1", ip)
}

func TestExtractClientIP_EmptyRemoteAddr(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = ""

	ip := extractClientIP(req)
	assert.Empty(t, ip)
}
