package jwt

import (
	"testing"
	"time"

	"github.com/Parallels/pd-api-service/basecontext"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func SetupToken(t *testing.T) *JwtSystemToken {
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.Options.WithSecret("secret")

	// Test case 1: Sign with valid input
	claimsMap := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	tokenStr, err := svc.Sign(claimsMap)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	token, err := svc.Parse(tokenStr)
	assert.NoError(t, err)

	return token
}

func TestValid(t *testing.T) {
	token := SetupToken(t)

	valid, err := token.Valid()
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestValidWithError(t *testing.T) {
	token := SetupToken(t)
	token.tokenObj = nil

	valid, err := token.Valid()
	assert.Errorf(t, err, "error: tokenObj is nil")
	assert.False(t, valid)
}

func TestValidExpired(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.Options.WithSecret("secret")
	svc.Options.WithTokenDuration(0.1)

	// Test case 1: Sign with valid input
	claimsMap := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	tokenStr, err := svc.Sign(claimsMap)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	token, err := svc.Parse(tokenStr)
	assert.NoError(t, err)
	time.Sleep(10 * time.Second)

	valid, err := token.Valid()
	assert.Errorf(t, err, "Token is expired")
	assert.False(t, valid)
}

func TestGetClaims(t *testing.T) {
	token := SetupToken(t)

	claims, err := token.GetTokenClaims()
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", claims["email"])
}

func TestGetEmail(t *testing.T) {
	token := SetupToken(t)

	email, err := token.GetEmail()
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", email)
}

func TestGetEmailNoClaim(t *testing.T) {
	token := SetupToken(t)
	token.Claims = nil
	token.tokenObj.Claims = nil

	email, err := token.GetEmail()
	assert.Errorf(t, err, "invalid claims")
	assert.Equal(t, "", email)
}

func TestGetEmailWrongFormat(t *testing.T) {
	token := SetupToken(t)
	token.Claims = nil
	token.tokenObj.Claims = jwt.MapClaims{
		"email": 2,
	}

	email, err := token.GetEmail()
	assert.Errorf(t, err, "invalid email")
	assert.Equal(t, "", email)
}

func TestGetExpiresAt(t *testing.T) {
	token := SetupToken(t)

	expiresAt, err := token.GetExpiresAt()
	assert.NoError(t, err)
	assert.True(t, time.Now().Before(expiresAt))
}

func TestGetExpiresAtNoClaim(t *testing.T) {
	token := SetupToken(t)
	token.Claims = nil
	token.tokenObj.Claims = nil

	expiresAt, err := token.GetExpiresAt()
	assert.Errorf(t, err, "invalid claims")
	assert.False(t, time.Now().Before(expiresAt))
}

func TestGetExpiresAtWrongFormat(t *testing.T) {
	token := SetupToken(t)
	token.Claims = nil
	token.tokenObj.Claims = jwt.MapClaims{
		"exp": "wrong",
	}

	expiresAt, err := token.GetExpiresAt()
	assert.Errorf(t, err, "invalid expiresAt")
	assert.False(t, time.Now().Before(expiresAt))
}

func TestGetClaim(t *testing.T) {
	ctx := basecontext.NewRootBaseContext()
	svc := New(ctx)
	svc.Options.WithSecret("secret")

	// Test case 1: Sign with valid input
	claimsMap := map[string]interface{}{
		"email":  "test@example.com",
		"uid":    "1234567890",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	tokenStr, err := svc.Sign(claimsMap)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	token, err := svc.Parse(tokenStr)
	assert.NoError(t, err)

	claim, err := token.GetClaim("uid")
	assert.NoError(t, err)
	assert.Equal(t, "1234567890", claim)
}

func TestGetClaimNoClaim(t *testing.T) {
	token := SetupToken(t)
	token.Claims = nil
	token.tokenObj.Claims = nil

	_, err := token.GetClaim("uid")
	assert.Errorf(t, err, "invalid claims")
}

func TestGetClaimWrongFormat(t *testing.T) {
	token := SetupToken(t)
	token.Claims = nil
	token.tokenObj.Claims = jwt.MapClaims{
		"exp": "wrong",
	}

	_, err := token.GetClaim("test")
	assert.Errorf(t, err, "invalid claim")
}
