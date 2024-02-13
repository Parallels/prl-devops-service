package jwt

import (
	"errors"
	"os"
	"testing"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/constants"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	globalJwtService = nil

	svc := Get()
	assert.NotNil(t, svc)

	assert.Equal(t, ctx, svc.ctx)

	svc2 := Get()
	assert.Equal(t, svc, svc2)
}

func TestJwtService_SignHS256(t *testing.T) {
	// Create a new instance of JwtService
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.WithSecret("secret")

	// Test case 1: Sign with valid input
	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	token, err := svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 2: Sign with empty email
	claims["email"] = ""
	token, err = svc.Sign(claims)
	assert.Error(t, err)
	assert.Empty(t, token)

	// Test case 3: Sign with empty roles
	claims["email"] = "test@example.com"
	claims["roles"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 4: Sign with empty claims
	claims["email"] = "test@example.com"
	claims["roles"] = []string{"admin", "user"}
	claims["claims"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJwtService_SignHS384(t *testing.T) {
	// Create a new instance of JwtService
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.WithSecret("secret")
	svc.WithAlgorithm(JwtSigningAlgorithmHS384)

	// Test case 1: Sign with valid input
	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	token, err := svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 2: Sign with empty email
	claims["email"] = ""
	token, err = svc.Sign(claims)
	assert.Error(t, err)
	assert.Empty(t, token)

	// Test case 3: Sign with empty roles
	claims["email"] = "test@example.com"
	claims["roles"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 4: Sign with empty claims
	claims["email"] = "test@example.com"
	claims["roles"] = []string{"admin", "user"}
	claims["claims"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJwtService_SignHS512(t *testing.T) {
	// Create a new instance of JwtService
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.WithSecret("secret")
	svc.WithAlgorithm(JwtSigningAlgorithmHS512)

	// Test case 1: Sign with valid input
	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	token, err := svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 2: Sign with empty email
	claims["email"] = ""
	token, err = svc.Sign(claims)
	assert.Error(t, err)
	assert.Empty(t, token)

	// Test case 3: Sign with empty roles
	claims["email"] = "test@example.com"
	claims["roles"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 4: Sign with empty claims
	claims["email"] = "test@example.com"
	claims["roles"] = []string{"admin", "user"}
	claims["claims"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJwtService_SignRandomSecret(t *testing.T) {
	// Create a new instance of JwtService
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)

	// Test case 1: Sign with valid input
	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	token, err := svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJwtService_SignRS256(t *testing.T) {
	// Create a new instance of JwtService
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.WithPrivateKey("LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBclJNKzFQaWIwb2VGQ0dPbzVtTmpzTjh4T3JML242SXQwTXZ1TzN3UmRwQ2V3SmZzClY3SVJwWW5MNU5MenRXZnVoNm9rbGd2cTIwak00NDUxQXpIM2tuZGt4NFhqY090Zmg5WklSby9xQlhYeEdGOVgKSldXem1zMjFqWkJzOTVaL3p3VFpncEJVL21WaEhSejVuOStVR1NlRVpZZnU4WnlMeEZGQ3JvQlY1Uk5rTnJIQwowSWJOT1ZuOFJ1Snprczg5dnl2UzZES2xLY29IZmppVVgwTE1hNUZxcStuRSs5R3kzUURzcWRrYUtwR0JUbmFrCnNlWkIraEV2TFBGWDFUMFRHZDhzQTQ4SzEyWjVZaklaNjBPRHppTnB5dktuL24zWW9qWjFMZno5QzFSbDYwM3EKb2VuYzF3Y3hZYis1WGxYempVZDdKU0ZxZlRPSHR4L21jb3J6RVFJREFRQUJBb0lCQUd6QkJDRlZHcXoreFNSUgpZMGRwdXJIZ2ZaR3grcGgza3l4NWlIaE9iTGJCQjFCajVubndoZEFzTnpmUktweVo4elBDUVAwYi9mYk8xeGhDCll0cWdJajRoMGV6aC85bnJ2UHFoTm9kSVZUZ2JOV1VvTzUxZk82MlJoM3l0L1JyU3Nmc0d4QnFiMFJ5TjNha3EKbE56VjQ3eXVyUzRUYlp6YXpxU01uMGNCazNlK1gvQ1dneTAvRGJ0WTRWZytXVVZybXRyNnY4cEpSYkE3ajRYago0ZG40MlNlbmdLaU1lU040SktjQ3JudkNMbG5lVDA4TDJOd2dNUjk2cFZzdldobGc4YUlVdnN5OUpsamw0UE5tClRQSXhOWHhEOUdQNUJacjBmSlgwRk94Mk1vdWh3TVNtcldBcmxta1gvTnprK1JkMGUvMm5RazF1M21ZMHIzS1MKcm9UckFSVUNnWUVBMmRCNW0xN2tOK0g5MHJON2ROWjlmTWYvdDdGRlY0THh1WUEyTXZ3WllDOFNGR2FNbWRneAoyMExHTkJrNER4OU16NHA4dDA2Q1kvOVQvMERqazlGOFVSbHJRSmw3dHpWTjhRcStGSGNuakFZOXpWL0RJTmZiCkFXb1E0WlFUM0k4c3lpTWJBSzh1T0s3ZWhYYlI5SWJhaWtmSmpKYzVkOVZnMkJOQUFDR1JMNjhDZ1lFQXkycmYKWnpVbHcySlNjUmhxOVQwQnpWdUZWU09wYWQxM3FLelZEczEzRW1Gb0dxazMxWnQxRW52YWl3VTd1R3UyYkdUVApnUUs2NVY5MEM1emdaVzA3eXoxQTd0U1ZHRURXZ2kzcll6bk1RekNVVUhYUng3Y1g5aTVHOE9uQkRLRnpuVnZHClhLU09GVnEzQVAyY09rcjN4bUYyNVNUKyt4d0cyQ2EwbFROQitUOENnWUVBb1hReUlDUkhvTkRJdUIxSXZ3T2IKQXhxeEI3WEVnNmpSaTBKcGFvT0tQOHpFWnhEWTJkVHlwK2VvU2NnRDBOblBzdXVocExMeVhqTk9UU0FKVVhIdgo1NkdpNmNDYmZ1TnBRZXBIbVozMVY0cnMxc1pNT3BVbWhyYmJpb3FiNmxyS3hZOGVIZlM4bTFHc0tsdzRKenlxCjArT0FsOUVrelJvQzdrZmVvZm8veDRzQ2dZQmE1WDNBbTZJdFJiRTdNa01SSk5xNlRnd3RlRXNLc0ZqNCtZb1gKSEQ3NTZxYmZTd0JWSml0UlFDRHZBRDZvY1JGS0xGL0toVkxJampmSHZLa1ZDWk92aE1hUU1sUVJTMS9QT2YrMgpEaXkxVlc3ZzZWVDlYbGFKdmpJYkV3a2R3TU50N0lXZC9qWXpXcDd1QldXYk1zYTNVZlFUL3MwbG5tZDhqUWNpCnFJM3hkd0tCZ1FDNmFiVHZzc20xRm53RFgzN0ZoS0NWeHpmYmpiY05MNmNmUG16ZXhiNFIxZmZBb2tJYlVEak0KV2xEZUExTndHMGlVRXdod0JCaXlrbFVUdnh6RkFLaG5rUDBHdllFa0Z4TXM1TFMwNjlITkxKcHhlMklFTXJSMwp4MW4veXQvV2w4T0RVYTM2S3czMGxwU0E4aE41UHR2NVpvRUVzWFA5L21aWWZTOHIzdCtjQkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ==")
	svc.WithAlgorithm(JwtSigningAlgorithmRS256)

	// Test case 1: Sign with valid input
	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	token, err := svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 2: Sign with empty email
	claims["email"] = ""
	token, err = svc.Sign(claims)
	assert.Error(t, err)
	assert.Empty(t, token)

	// Test case 3: Sign with empty roles
	claims["email"] = "test@example.com"
	claims["roles"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 4: Sign with empty claims
	claims["email"] = "test@example.com"
	claims["roles"] = []string{"admin", "user"}
	claims["claims"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJwtService_SignRS384(t *testing.T) {
	// Create a new instance of JwtService
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.WithPrivateKey("LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBclJNKzFQaWIwb2VGQ0dPbzVtTmpzTjh4T3JML242SXQwTXZ1TzN3UmRwQ2V3SmZzClY3SVJwWW5MNU5MenRXZnVoNm9rbGd2cTIwak00NDUxQXpIM2tuZGt4NFhqY090Zmg5WklSby9xQlhYeEdGOVgKSldXem1zMjFqWkJzOTVaL3p3VFpncEJVL21WaEhSejVuOStVR1NlRVpZZnU4WnlMeEZGQ3JvQlY1Uk5rTnJIQwowSWJOT1ZuOFJ1Snprczg5dnl2UzZES2xLY29IZmppVVgwTE1hNUZxcStuRSs5R3kzUURzcWRrYUtwR0JUbmFrCnNlWkIraEV2TFBGWDFUMFRHZDhzQTQ4SzEyWjVZaklaNjBPRHppTnB5dktuL24zWW9qWjFMZno5QzFSbDYwM3EKb2VuYzF3Y3hZYis1WGxYempVZDdKU0ZxZlRPSHR4L21jb3J6RVFJREFRQUJBb0lCQUd6QkJDRlZHcXoreFNSUgpZMGRwdXJIZ2ZaR3grcGgza3l4NWlIaE9iTGJCQjFCajVubndoZEFzTnpmUktweVo4elBDUVAwYi9mYk8xeGhDCll0cWdJajRoMGV6aC85bnJ2UHFoTm9kSVZUZ2JOV1VvTzUxZk82MlJoM3l0L1JyU3Nmc0d4QnFiMFJ5TjNha3EKbE56VjQ3eXVyUzRUYlp6YXpxU01uMGNCazNlK1gvQ1dneTAvRGJ0WTRWZytXVVZybXRyNnY4cEpSYkE3ajRYago0ZG40MlNlbmdLaU1lU040SktjQ3JudkNMbG5lVDA4TDJOd2dNUjk2cFZzdldobGc4YUlVdnN5OUpsamw0UE5tClRQSXhOWHhEOUdQNUJacjBmSlgwRk94Mk1vdWh3TVNtcldBcmxta1gvTnprK1JkMGUvMm5RazF1M21ZMHIzS1MKcm9UckFSVUNnWUVBMmRCNW0xN2tOK0g5MHJON2ROWjlmTWYvdDdGRlY0THh1WUEyTXZ3WllDOFNGR2FNbWRneAoyMExHTkJrNER4OU16NHA4dDA2Q1kvOVQvMERqazlGOFVSbHJRSmw3dHpWTjhRcStGSGNuakFZOXpWL0RJTmZiCkFXb1E0WlFUM0k4c3lpTWJBSzh1T0s3ZWhYYlI5SWJhaWtmSmpKYzVkOVZnMkJOQUFDR1JMNjhDZ1lFQXkycmYKWnpVbHcySlNjUmhxOVQwQnpWdUZWU09wYWQxM3FLelZEczEzRW1Gb0dxazMxWnQxRW52YWl3VTd1R3UyYkdUVApnUUs2NVY5MEM1emdaVzA3eXoxQTd0U1ZHRURXZ2kzcll6bk1RekNVVUhYUng3Y1g5aTVHOE9uQkRLRnpuVnZHClhLU09GVnEzQVAyY09rcjN4bUYyNVNUKyt4d0cyQ2EwbFROQitUOENnWUVBb1hReUlDUkhvTkRJdUIxSXZ3T2IKQXhxeEI3WEVnNmpSaTBKcGFvT0tQOHpFWnhEWTJkVHlwK2VvU2NnRDBOblBzdXVocExMeVhqTk9UU0FKVVhIdgo1NkdpNmNDYmZ1TnBRZXBIbVozMVY0cnMxc1pNT3BVbWhyYmJpb3FiNmxyS3hZOGVIZlM4bTFHc0tsdzRKenlxCjArT0FsOUVrelJvQzdrZmVvZm8veDRzQ2dZQmE1WDNBbTZJdFJiRTdNa01SSk5xNlRnd3RlRXNLc0ZqNCtZb1gKSEQ3NTZxYmZTd0JWSml0UlFDRHZBRDZvY1JGS0xGL0toVkxJampmSHZLa1ZDWk92aE1hUU1sUVJTMS9QT2YrMgpEaXkxVlc3ZzZWVDlYbGFKdmpJYkV3a2R3TU50N0lXZC9qWXpXcDd1QldXYk1zYTNVZlFUL3MwbG5tZDhqUWNpCnFJM3hkd0tCZ1FDNmFiVHZzc20xRm53RFgzN0ZoS0NWeHpmYmpiY05MNmNmUG16ZXhiNFIxZmZBb2tJYlVEak0KV2xEZUExTndHMGlVRXdod0JCaXlrbFVUdnh6RkFLaG5rUDBHdllFa0Z4TXM1TFMwNjlITkxKcHhlMklFTXJSMwp4MW4veXQvV2w4T0RVYTM2S3czMGxwU0E4aE41UHR2NVpvRUVzWFA5L21aWWZTOHIzdCtjQkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ==")
	svc.WithAlgorithm(JwtSigningAlgorithmRS384)

	// Test case 1: Sign with valid input
	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	token, err := svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 2: Sign with empty email
	claims["email"] = ""
	token, err = svc.Sign(claims)
	assert.Error(t, err)
	assert.Empty(t, token)

	// Test case 3: Sign with empty roles
	claims["email"] = "test@example.com"
	claims["roles"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 4: Sign with empty claims
	claims["email"] = "test@example.com"
	claims["roles"] = []string{"admin", "user"}
	claims["claims"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJwtService_SignRS512(t *testing.T) {
	// Create a new instance of JwtService
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.WithPrivateKey("LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBclJNKzFQaWIwb2VGQ0dPbzVtTmpzTjh4T3JML242SXQwTXZ1TzN3UmRwQ2V3SmZzClY3SVJwWW5MNU5MenRXZnVoNm9rbGd2cTIwak00NDUxQXpIM2tuZGt4NFhqY090Zmg5WklSby9xQlhYeEdGOVgKSldXem1zMjFqWkJzOTVaL3p3VFpncEJVL21WaEhSejVuOStVR1NlRVpZZnU4WnlMeEZGQ3JvQlY1Uk5rTnJIQwowSWJOT1ZuOFJ1Snprczg5dnl2UzZES2xLY29IZmppVVgwTE1hNUZxcStuRSs5R3kzUURzcWRrYUtwR0JUbmFrCnNlWkIraEV2TFBGWDFUMFRHZDhzQTQ4SzEyWjVZaklaNjBPRHppTnB5dktuL24zWW9qWjFMZno5QzFSbDYwM3EKb2VuYzF3Y3hZYis1WGxYempVZDdKU0ZxZlRPSHR4L21jb3J6RVFJREFRQUJBb0lCQUd6QkJDRlZHcXoreFNSUgpZMGRwdXJIZ2ZaR3grcGgza3l4NWlIaE9iTGJCQjFCajVubndoZEFzTnpmUktweVo4elBDUVAwYi9mYk8xeGhDCll0cWdJajRoMGV6aC85bnJ2UHFoTm9kSVZUZ2JOV1VvTzUxZk82MlJoM3l0L1JyU3Nmc0d4QnFiMFJ5TjNha3EKbE56VjQ3eXVyUzRUYlp6YXpxU01uMGNCazNlK1gvQ1dneTAvRGJ0WTRWZytXVVZybXRyNnY4cEpSYkE3ajRYago0ZG40MlNlbmdLaU1lU040SktjQ3JudkNMbG5lVDA4TDJOd2dNUjk2cFZzdldobGc4YUlVdnN5OUpsamw0UE5tClRQSXhOWHhEOUdQNUJacjBmSlgwRk94Mk1vdWh3TVNtcldBcmxta1gvTnprK1JkMGUvMm5RazF1M21ZMHIzS1MKcm9UckFSVUNnWUVBMmRCNW0xN2tOK0g5MHJON2ROWjlmTWYvdDdGRlY0THh1WUEyTXZ3WllDOFNGR2FNbWRneAoyMExHTkJrNER4OU16NHA4dDA2Q1kvOVQvMERqazlGOFVSbHJRSmw3dHpWTjhRcStGSGNuakFZOXpWL0RJTmZiCkFXb1E0WlFUM0k4c3lpTWJBSzh1T0s3ZWhYYlI5SWJhaWtmSmpKYzVkOVZnMkJOQUFDR1JMNjhDZ1lFQXkycmYKWnpVbHcySlNjUmhxOVQwQnpWdUZWU09wYWQxM3FLelZEczEzRW1Gb0dxazMxWnQxRW52YWl3VTd1R3UyYkdUVApnUUs2NVY5MEM1emdaVzA3eXoxQTd0U1ZHRURXZ2kzcll6bk1RekNVVUhYUng3Y1g5aTVHOE9uQkRLRnpuVnZHClhLU09GVnEzQVAyY09rcjN4bUYyNVNUKyt4d0cyQ2EwbFROQitUOENnWUVBb1hReUlDUkhvTkRJdUIxSXZ3T2IKQXhxeEI3WEVnNmpSaTBKcGFvT0tQOHpFWnhEWTJkVHlwK2VvU2NnRDBOblBzdXVocExMeVhqTk9UU0FKVVhIdgo1NkdpNmNDYmZ1TnBRZXBIbVozMVY0cnMxc1pNT3BVbWhyYmJpb3FiNmxyS3hZOGVIZlM4bTFHc0tsdzRKenlxCjArT0FsOUVrelJvQzdrZmVvZm8veDRzQ2dZQmE1WDNBbTZJdFJiRTdNa01SSk5xNlRnd3RlRXNLc0ZqNCtZb1gKSEQ3NTZxYmZTd0JWSml0UlFDRHZBRDZvY1JGS0xGL0toVkxJampmSHZLa1ZDWk92aE1hUU1sUVJTMS9QT2YrMgpEaXkxVlc3ZzZWVDlYbGFKdmpJYkV3a2R3TU50N0lXZC9qWXpXcDd1QldXYk1zYTNVZlFUL3MwbG5tZDhqUWNpCnFJM3hkd0tCZ1FDNmFiVHZzc20xRm53RFgzN0ZoS0NWeHpmYmpiY05MNmNmUG16ZXhiNFIxZmZBb2tJYlVEak0KV2xEZUExTndHMGlVRXdod0JCaXlrbFVUdnh6RkFLaG5rUDBHdllFa0Z4TXM1TFMwNjlITkxKcHhlMklFTXJSMwp4MW4veXQvV2w4T0RVYTM2S3czMGxwU0E4aE41UHR2NVpvRUVzWFA5L21aWWZTOHIzdCtjQkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ==")
	svc.WithAlgorithm(JwtSigningAlgorithmRS512)

	// Test case 1: Sign with valid input
	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	token, err := svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 2: Sign with empty email
	claims["email"] = ""
	token, err = svc.Sign(claims)
	assert.Error(t, err)
	assert.Empty(t, token)

	// Test case 3: Sign with empty roles
	claims["email"] = "test@example.com"
	claims["roles"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 4: Sign with empty claims
	claims["email"] = "test@example.com"
	claims["roles"] = []string{"admin", "user"}
	claims["claims"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJwtService_SignDefault(t *testing.T) {
	// Create a new instance of JwtService
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.WithSecret("secret")
	svc.WithAlgorithm("")

	// Test case 1: Sign with valid input
	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	token, err := svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 2: Sign with empty email
	claims["email"] = ""
	token, err = svc.Sign(claims)
	assert.Error(t, err)
	assert.Empty(t, token)

	// Test case 3: Sign with empty roles
	claims["email"] = "test@example.com"
	claims["roles"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test case 4: Sign with empty claims
	claims["email"] = "test@example.com"
	claims["roles"] = []string{"admin", "user"}
	claims["claims"] = []string{}
	token, err = svc.Sign(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJwtService_SignNoPrivateKey(t *testing.T) {
	// Create a new instance of JwtService
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.WithPrivateKey("")
	svc.WithAlgorithm(JwtSigningAlgorithmRS256)

	// Test case 1: Sign with valid input
	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	_, err := svc.Sign(claims)
	assert.Errorf(t, err, "private key cannot be empty")
}

func TestGenerateJwksRS256Algorithm(t *testing.T) {
	svc := New(nil)
	svc.Options.PrivateKey = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBclJNKzFQaWIwb2VGQ0dPbzVtTmpzTjh4T3JML242SXQwTXZ1TzN3UmRwQ2V3SmZzClY3SVJwWW5MNU5MenRXZnVoNm9rbGd2cTIwak00NDUxQXpIM2tuZGt4NFhqY090Zmg5WklSby9xQlhYeEdGOVgKSldXem1zMjFqWkJzOTVaL3p3VFpncEJVL21WaEhSejVuOStVR1NlRVpZZnU4WnlMeEZGQ3JvQlY1Uk5rTnJIQwowSWJOT1ZuOFJ1Snprczg5dnl2UzZES2xLY29IZmppVVgwTE1hNUZxcStuRSs5R3kzUURzcWRrYUtwR0JUbmFrCnNlWkIraEV2TFBGWDFUMFRHZDhzQTQ4SzEyWjVZaklaNjBPRHppTnB5dktuL24zWW9qWjFMZno5QzFSbDYwM3EKb2VuYzF3Y3hZYis1WGxYempVZDdKU0ZxZlRPSHR4L21jb3J6RVFJREFRQUJBb0lCQUd6QkJDRlZHcXoreFNSUgpZMGRwdXJIZ2ZaR3grcGgza3l4NWlIaE9iTGJCQjFCajVubndoZEFzTnpmUktweVo4elBDUVAwYi9mYk8xeGhDCll0cWdJajRoMGV6aC85bnJ2UHFoTm9kSVZUZ2JOV1VvTzUxZk82MlJoM3l0L1JyU3Nmc0d4QnFiMFJ5TjNha3EKbE56VjQ3eXVyUzRUYlp6YXpxU01uMGNCazNlK1gvQ1dneTAvRGJ0WTRWZytXVVZybXRyNnY4cEpSYkE3ajRYago0ZG40MlNlbmdLaU1lU040SktjQ3JudkNMbG5lVDA4TDJOd2dNUjk2cFZzdldobGc4YUlVdnN5OUpsamw0UE5tClRQSXhOWHhEOUdQNUJacjBmSlgwRk94Mk1vdWh3TVNtcldBcmxta1gvTnprK1JkMGUvMm5RazF1M21ZMHIzS1MKcm9UckFSVUNnWUVBMmRCNW0xN2tOK0g5MHJON2ROWjlmTWYvdDdGRlY0THh1WUEyTXZ3WllDOFNGR2FNbWRneAoyMExHTkJrNER4OU16NHA4dDA2Q1kvOVQvMERqazlGOFVSbHJRSmw3dHpWTjhRcStGSGNuakFZOXpWL0RJTmZiCkFXb1E0WlFUM0k4c3lpTWJBSzh1T0s3ZWhYYlI5SWJhaWtmSmpKYzVkOVZnMkJOQUFDR1JMNjhDZ1lFQXkycmYKWnpVbHcySlNjUmhxOVQwQnpWdUZWU09wYWQxM3FLelZEczEzRW1Gb0dxazMxWnQxRW52YWl3VTd1R3UyYkdUVApnUUs2NVY5MEM1emdaVzA3eXoxQTd0U1ZHRURXZ2kzcll6bk1RekNVVUhYUng3Y1g5aTVHOE9uQkRLRnpuVnZHClhLU09GVnEzQVAyY09rcjN4bUYyNVNUKyt4d0cyQ2EwbFROQitUOENnWUVBb1hReUlDUkhvTkRJdUIxSXZ3T2IKQXhxeEI3WEVnNmpSaTBKcGFvT0tQOHpFWnhEWTJkVHlwK2VvU2NnRDBOblBzdXVocExMeVhqTk9UU0FKVVhIdgo1NkdpNmNDYmZ1TnBRZXBIbVozMVY0cnMxc1pNT3BVbWhyYmJpb3FiNmxyS3hZOGVIZlM4bTFHc0tsdzRKenlxCjArT0FsOUVrelJvQzdrZmVvZm8veDRzQ2dZQmE1WDNBbTZJdFJiRTdNa01SSk5xNlRnd3RlRXNLc0ZqNCtZb1gKSEQ3NTZxYmZTd0JWSml0UlFDRHZBRDZvY1JGS0xGL0toVkxJampmSHZLa1ZDWk92aE1hUU1sUVJTMS9QT2YrMgpEaXkxVlc3ZzZWVDlYbGFKdmpJYkV3a2R3TU50N0lXZC9qWXpXcDd1QldXYk1zYTNVZlFUL3MwbG5tZDhqUWNpCnFJM3hkd0tCZ1FDNmFiVHZzc20xRm53RFgzN0ZoS0NWeHpmYmpiY05MNmNmUG16ZXhiNFIxZmZBb2tJYlVEak0KV2xEZUExTndHMGlVRXdod0JCaXlrbFVUdnh6RkFLaG5rUDBHdllFa0Z4TXM1TFMwNjlITkxKcHhlMklFTXJSMwp4MW4veXQvV2w4T0RVYTM2S3czMGxwU0E4aE41UHR2NVpvRUVzWFA5L21aWWZTOHIzdCtjQkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ=="
	svc.Options.Algorithm = JwtSigningAlgorithmRS256

	expectedJWK := `{"kty":"RSA","kid":"ee8146d4b30a57d0053f39c80f4c3caa46461633","alg":"RS256","n":"rRM-1Pib0oeFCGOo5mNjsN8xOrL_n6It0MvuO3wRdpCewJfsV7IRpYnL5NLztWfuh6oklgvq20jM4451AzH3kndkx4XjcOtfh9ZIRo_qBXXxGF9XJWWzms21jZBs95Z_zwTZgpBU_mVhHRz5n9-UGSeEZYfu8ZyLxFFCroBV5RNkNrHC0IbNOVn8RuJzks89vyvS6DKlKcoHfjiUX0LMa5Fqq-nE-9Gy3QDsqdkaKpGBTnakseZB-hEvLPFX1T0TGd8sA48K12Z5YjIZ60ODziNpyvKn_n3YojZ1Lfz9C1Rl603qoenc1wcxYb-5XlXzjUd7JSFqfTOHtx_mcorzEQ","e":"AQAB","d":"bMEEIVUarP7FJFFjR2m6seB9kbH6mHeTLHmIeE5stsEHUGPmefCF0Cw3N9EqnJnzM8JA_Rv99s7XGEJi2qAiPiHR7OH_2eu8-qE2h0hVOBs1ZSg7nV87rZGHfK39GtKx-wbEGpvRHI3dqSqU3NXjvK6tLhNtnNrOpIyfRwGTd75f8JaDLT8Nu1jhWD5ZRWua2vq_yklFsDuPhePh2fjZJ6eAqIx5I3gkpwKue8IuWd5PTwvY3CAxH3qlWy9aGWDxohS-zL0mWOXg82ZM8jE1fEP0Y_kFmvR8lfQU7HYyi6HAxKatYCuWaRf83OT5F3R7_adCTW7eZjSvcpKuhOsBFQ","p":"2dB5m17kN-H90rN7dNZ9fMf_t7FFV4LxuYA2MvwZYC8SFGaMmdgx20LGNBk4Dx9Mz4p8t06CY_9T_0Djk9F8URlrQJl7tzVN8Qq-FHcnjAY9zV_DINfbAWoQ4ZQT3I8syiMbAK8uOK7ehXbR9IbaikfJjJc5d9Vg2BNAACGRL68","q":"y2rfZzUlw2JScRhq9T0BzVuFVSOpad13qKzVDs13EmFoGqk31Zt1EnvaiwU7uGu2bGTTgQK65V90C5zgZW07yz1A7tSVGEDWgi3rYznMQzCUUHXRx7cX9i5G8OnBDKFznVvGXKSOFVq3AP2cOkr3xmF25ST--xwG2Ca0lTNB-T8","dp":"oXQyICRHoNDIuB1IvwObAxqxB7XEg6jRi0JpaoOKP8zEZxDY2dTyp-eoScgD0NnPsuuhpLLyXjNOTSAJUXHv56Gi6cCbfuNpQepHmZ31V4rs1sZMOpUmhrbbioqb6lrKxY8eHfS8m1GsKlw4Jzyq0-OAl9EkzRoC7kfeofo_x4s","dq":"WuV9wJuiLUWxOzJDESTauk4MLXhLCrBY-PmKFxw--eqm30sAVSYrUUAg7wA-qHERSixfyoVSyI43x7ypFQmTr4TGkDJUEUtfzzn_tg4stVVu4OlU_V5Wib4yGxMJHcDDbeyFnf42M1qe7gVlmzLGt1H0E_7NJZ5nfI0HIqiN8Xc","qi":"umm077LJtRZ8A19-xYSglcc32423DS-nHz5s3sW-EdX3wKJCG1A4zFpQ3gNTcBtIlBMIcAQYspJVE78cxQCoZ5D9Br2BJBcTLOS0tOvRzSyacXtiBDK0d8dZ_8rf1pfDg1Gt-isN9JaUgPITeT7b-WaBBLFz_f5mWH0vK97fnAQ"}`

	jwk, err := svc.GenerateJWKS()
	assert.NoError(t, err)
	assert.Equal(t, expectedJWK, jwk)
}

func TestGenerateJwksRS384Algorithm(t *testing.T) {
	svc := New(nil)
	svc.Options.PrivateKey = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBclJNKzFQaWIwb2VGQ0dPbzVtTmpzTjh4T3JML242SXQwTXZ1TzN3UmRwQ2V3SmZzClY3SVJwWW5MNU5MenRXZnVoNm9rbGd2cTIwak00NDUxQXpIM2tuZGt4NFhqY090Zmg5WklSby9xQlhYeEdGOVgKSldXem1zMjFqWkJzOTVaL3p3VFpncEJVL21WaEhSejVuOStVR1NlRVpZZnU4WnlMeEZGQ3JvQlY1Uk5rTnJIQwowSWJOT1ZuOFJ1Snprczg5dnl2UzZES2xLY29IZmppVVgwTE1hNUZxcStuRSs5R3kzUURzcWRrYUtwR0JUbmFrCnNlWkIraEV2TFBGWDFUMFRHZDhzQTQ4SzEyWjVZaklaNjBPRHppTnB5dktuL24zWW9qWjFMZno5QzFSbDYwM3EKb2VuYzF3Y3hZYis1WGxYempVZDdKU0ZxZlRPSHR4L21jb3J6RVFJREFRQUJBb0lCQUd6QkJDRlZHcXoreFNSUgpZMGRwdXJIZ2ZaR3grcGgza3l4NWlIaE9iTGJCQjFCajVubndoZEFzTnpmUktweVo4elBDUVAwYi9mYk8xeGhDCll0cWdJajRoMGV6aC85bnJ2UHFoTm9kSVZUZ2JOV1VvTzUxZk82MlJoM3l0L1JyU3Nmc0d4QnFiMFJ5TjNha3EKbE56VjQ3eXVyUzRUYlp6YXpxU01uMGNCazNlK1gvQ1dneTAvRGJ0WTRWZytXVVZybXRyNnY4cEpSYkE3ajRYago0ZG40MlNlbmdLaU1lU040SktjQ3JudkNMbG5lVDA4TDJOd2dNUjk2cFZzdldobGc4YUlVdnN5OUpsamw0UE5tClRQSXhOWHhEOUdQNUJacjBmSlgwRk94Mk1vdWh3TVNtcldBcmxta1gvTnprK1JkMGUvMm5RazF1M21ZMHIzS1MKcm9UckFSVUNnWUVBMmRCNW0xN2tOK0g5MHJON2ROWjlmTWYvdDdGRlY0THh1WUEyTXZ3WllDOFNGR2FNbWRneAoyMExHTkJrNER4OU16NHA4dDA2Q1kvOVQvMERqazlGOFVSbHJRSmw3dHpWTjhRcStGSGNuakFZOXpWL0RJTmZiCkFXb1E0WlFUM0k4c3lpTWJBSzh1T0s3ZWhYYlI5SWJhaWtmSmpKYzVkOVZnMkJOQUFDR1JMNjhDZ1lFQXkycmYKWnpVbHcySlNjUmhxOVQwQnpWdUZWU09wYWQxM3FLelZEczEzRW1Gb0dxazMxWnQxRW52YWl3VTd1R3UyYkdUVApnUUs2NVY5MEM1emdaVzA3eXoxQTd0U1ZHRURXZ2kzcll6bk1RekNVVUhYUng3Y1g5aTVHOE9uQkRLRnpuVnZHClhLU09GVnEzQVAyY09rcjN4bUYyNVNUKyt4d0cyQ2EwbFROQitUOENnWUVBb1hReUlDUkhvTkRJdUIxSXZ3T2IKQXhxeEI3WEVnNmpSaTBKcGFvT0tQOHpFWnhEWTJkVHlwK2VvU2NnRDBOblBzdXVocExMeVhqTk9UU0FKVVhIdgo1NkdpNmNDYmZ1TnBRZXBIbVozMVY0cnMxc1pNT3BVbWhyYmJpb3FiNmxyS3hZOGVIZlM4bTFHc0tsdzRKenlxCjArT0FsOUVrelJvQzdrZmVvZm8veDRzQ2dZQmE1WDNBbTZJdFJiRTdNa01SSk5xNlRnd3RlRXNLc0ZqNCtZb1gKSEQ3NTZxYmZTd0JWSml0UlFDRHZBRDZvY1JGS0xGL0toVkxJampmSHZLa1ZDWk92aE1hUU1sUVJTMS9QT2YrMgpEaXkxVlc3ZzZWVDlYbGFKdmpJYkV3a2R3TU50N0lXZC9qWXpXcDd1QldXYk1zYTNVZlFUL3MwbG5tZDhqUWNpCnFJM3hkd0tCZ1FDNmFiVHZzc20xRm53RFgzN0ZoS0NWeHpmYmpiY05MNmNmUG16ZXhiNFIxZmZBb2tJYlVEak0KV2xEZUExTndHMGlVRXdod0JCaXlrbFVUdnh6RkFLaG5rUDBHdllFa0Z4TXM1TFMwNjlITkxKcHhlMklFTXJSMwp4MW4veXQvV2w4T0RVYTM2S3czMGxwU0E4aE41UHR2NVpvRUVzWFA5L21aWWZTOHIzdCtjQkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ=="
	svc.Options.Algorithm = JwtSigningAlgorithmRS384

	expectedJWK := `{"kty":"RSA","kid":"ee8146d4b30a57d0053f39c80f4c3caa46461633","alg":"RS384","n":"rRM-1Pib0oeFCGOo5mNjsN8xOrL_n6It0MvuO3wRdpCewJfsV7IRpYnL5NLztWfuh6oklgvq20jM4451AzH3kndkx4XjcOtfh9ZIRo_qBXXxGF9XJWWzms21jZBs95Z_zwTZgpBU_mVhHRz5n9-UGSeEZYfu8ZyLxFFCroBV5RNkNrHC0IbNOVn8RuJzks89vyvS6DKlKcoHfjiUX0LMa5Fqq-nE-9Gy3QDsqdkaKpGBTnakseZB-hEvLPFX1T0TGd8sA48K12Z5YjIZ60ODziNpyvKn_n3YojZ1Lfz9C1Rl603qoenc1wcxYb-5XlXzjUd7JSFqfTOHtx_mcorzEQ","e":"AQAB","d":"bMEEIVUarP7FJFFjR2m6seB9kbH6mHeTLHmIeE5stsEHUGPmefCF0Cw3N9EqnJnzM8JA_Rv99s7XGEJi2qAiPiHR7OH_2eu8-qE2h0hVOBs1ZSg7nV87rZGHfK39GtKx-wbEGpvRHI3dqSqU3NXjvK6tLhNtnNrOpIyfRwGTd75f8JaDLT8Nu1jhWD5ZRWua2vq_yklFsDuPhePh2fjZJ6eAqIx5I3gkpwKue8IuWd5PTwvY3CAxH3qlWy9aGWDxohS-zL0mWOXg82ZM8jE1fEP0Y_kFmvR8lfQU7HYyi6HAxKatYCuWaRf83OT5F3R7_adCTW7eZjSvcpKuhOsBFQ","p":"2dB5m17kN-H90rN7dNZ9fMf_t7FFV4LxuYA2MvwZYC8SFGaMmdgx20LGNBk4Dx9Mz4p8t06CY_9T_0Djk9F8URlrQJl7tzVN8Qq-FHcnjAY9zV_DINfbAWoQ4ZQT3I8syiMbAK8uOK7ehXbR9IbaikfJjJc5d9Vg2BNAACGRL68","q":"y2rfZzUlw2JScRhq9T0BzVuFVSOpad13qKzVDs13EmFoGqk31Zt1EnvaiwU7uGu2bGTTgQK65V90C5zgZW07yz1A7tSVGEDWgi3rYznMQzCUUHXRx7cX9i5G8OnBDKFznVvGXKSOFVq3AP2cOkr3xmF25ST--xwG2Ca0lTNB-T8","dp":"oXQyICRHoNDIuB1IvwObAxqxB7XEg6jRi0JpaoOKP8zEZxDY2dTyp-eoScgD0NnPsuuhpLLyXjNOTSAJUXHv56Gi6cCbfuNpQepHmZ31V4rs1sZMOpUmhrbbioqb6lrKxY8eHfS8m1GsKlw4Jzyq0-OAl9EkzRoC7kfeofo_x4s","dq":"WuV9wJuiLUWxOzJDESTauk4MLXhLCrBY-PmKFxw--eqm30sAVSYrUUAg7wA-qHERSixfyoVSyI43x7ypFQmTr4TGkDJUEUtfzzn_tg4stVVu4OlU_V5Wib4yGxMJHcDDbeyFnf42M1qe7gVlmzLGt1H0E_7NJZ5nfI0HIqiN8Xc","qi":"umm077LJtRZ8A19-xYSglcc32423DS-nHz5s3sW-EdX3wKJCG1A4zFpQ3gNTcBtIlBMIcAQYspJVE78cxQCoZ5D9Br2BJBcTLOS0tOvRzSyacXtiBDK0d8dZ_8rf1pfDg1Gt-isN9JaUgPITeT7b-WaBBLFz_f5mWH0vK97fnAQ"}`

	jwk, err := svc.GenerateJWKS()
	assert.NoError(t, err)
	assert.Equal(t, expectedJWK, jwk)
}

func TestGenerateJwksRS512Algorithm(t *testing.T) {
	svc := New(nil)
	svc.Options.PrivateKey = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBclJNKzFQaWIwb2VGQ0dPbzVtTmpzTjh4T3JML242SXQwTXZ1TzN3UmRwQ2V3SmZzClY3SVJwWW5MNU5MenRXZnVoNm9rbGd2cTIwak00NDUxQXpIM2tuZGt4NFhqY090Zmg5WklSby9xQlhYeEdGOVgKSldXem1zMjFqWkJzOTVaL3p3VFpncEJVL21WaEhSejVuOStVR1NlRVpZZnU4WnlMeEZGQ3JvQlY1Uk5rTnJIQwowSWJOT1ZuOFJ1Snprczg5dnl2UzZES2xLY29IZmppVVgwTE1hNUZxcStuRSs5R3kzUURzcWRrYUtwR0JUbmFrCnNlWkIraEV2TFBGWDFUMFRHZDhzQTQ4SzEyWjVZaklaNjBPRHppTnB5dktuL24zWW9qWjFMZno5QzFSbDYwM3EKb2VuYzF3Y3hZYis1WGxYempVZDdKU0ZxZlRPSHR4L21jb3J6RVFJREFRQUJBb0lCQUd6QkJDRlZHcXoreFNSUgpZMGRwdXJIZ2ZaR3grcGgza3l4NWlIaE9iTGJCQjFCajVubndoZEFzTnpmUktweVo4elBDUVAwYi9mYk8xeGhDCll0cWdJajRoMGV6aC85bnJ2UHFoTm9kSVZUZ2JOV1VvTzUxZk82MlJoM3l0L1JyU3Nmc0d4QnFiMFJ5TjNha3EKbE56VjQ3eXVyUzRUYlp6YXpxU01uMGNCazNlK1gvQ1dneTAvRGJ0WTRWZytXVVZybXRyNnY4cEpSYkE3ajRYago0ZG40MlNlbmdLaU1lU040SktjQ3JudkNMbG5lVDA4TDJOd2dNUjk2cFZzdldobGc4YUlVdnN5OUpsamw0UE5tClRQSXhOWHhEOUdQNUJacjBmSlgwRk94Mk1vdWh3TVNtcldBcmxta1gvTnprK1JkMGUvMm5RazF1M21ZMHIzS1MKcm9UckFSVUNnWUVBMmRCNW0xN2tOK0g5MHJON2ROWjlmTWYvdDdGRlY0THh1WUEyTXZ3WllDOFNGR2FNbWRneAoyMExHTkJrNER4OU16NHA4dDA2Q1kvOVQvMERqazlGOFVSbHJRSmw3dHpWTjhRcStGSGNuakFZOXpWL0RJTmZiCkFXb1E0WlFUM0k4c3lpTWJBSzh1T0s3ZWhYYlI5SWJhaWtmSmpKYzVkOVZnMkJOQUFDR1JMNjhDZ1lFQXkycmYKWnpVbHcySlNjUmhxOVQwQnpWdUZWU09wYWQxM3FLelZEczEzRW1Gb0dxazMxWnQxRW52YWl3VTd1R3UyYkdUVApnUUs2NVY5MEM1emdaVzA3eXoxQTd0U1ZHRURXZ2kzcll6bk1RekNVVUhYUng3Y1g5aTVHOE9uQkRLRnpuVnZHClhLU09GVnEzQVAyY09rcjN4bUYyNVNUKyt4d0cyQ2EwbFROQitUOENnWUVBb1hReUlDUkhvTkRJdUIxSXZ3T2IKQXhxeEI3WEVnNmpSaTBKcGFvT0tQOHpFWnhEWTJkVHlwK2VvU2NnRDBOblBzdXVocExMeVhqTk9UU0FKVVhIdgo1NkdpNmNDYmZ1TnBRZXBIbVozMVY0cnMxc1pNT3BVbWhyYmJpb3FiNmxyS3hZOGVIZlM4bTFHc0tsdzRKenlxCjArT0FsOUVrelJvQzdrZmVvZm8veDRzQ2dZQmE1WDNBbTZJdFJiRTdNa01SSk5xNlRnd3RlRXNLc0ZqNCtZb1gKSEQ3NTZxYmZTd0JWSml0UlFDRHZBRDZvY1JGS0xGL0toVkxJampmSHZLa1ZDWk92aE1hUU1sUVJTMS9QT2YrMgpEaXkxVlc3ZzZWVDlYbGFKdmpJYkV3a2R3TU50N0lXZC9qWXpXcDd1QldXYk1zYTNVZlFUL3MwbG5tZDhqUWNpCnFJM3hkd0tCZ1FDNmFiVHZzc20xRm53RFgzN0ZoS0NWeHpmYmpiY05MNmNmUG16ZXhiNFIxZmZBb2tJYlVEak0KV2xEZUExTndHMGlVRXdod0JCaXlrbFVUdnh6RkFLaG5rUDBHdllFa0Z4TXM1TFMwNjlITkxKcHhlMklFTXJSMwp4MW4veXQvV2w4T0RVYTM2S3czMGxwU0E4aE41UHR2NVpvRUVzWFA5L21aWWZTOHIzdCtjQkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ=="
	svc.Options.Algorithm = JwtSigningAlgorithmRS512

	expectedJWK := `{"kty":"RSA","kid":"ee8146d4b30a57d0053f39c80f4c3caa46461633","alg":"RS512","n":"rRM-1Pib0oeFCGOo5mNjsN8xOrL_n6It0MvuO3wRdpCewJfsV7IRpYnL5NLztWfuh6oklgvq20jM4451AzH3kndkx4XjcOtfh9ZIRo_qBXXxGF9XJWWzms21jZBs95Z_zwTZgpBU_mVhHRz5n9-UGSeEZYfu8ZyLxFFCroBV5RNkNrHC0IbNOVn8RuJzks89vyvS6DKlKcoHfjiUX0LMa5Fqq-nE-9Gy3QDsqdkaKpGBTnakseZB-hEvLPFX1T0TGd8sA48K12Z5YjIZ60ODziNpyvKn_n3YojZ1Lfz9C1Rl603qoenc1wcxYb-5XlXzjUd7JSFqfTOHtx_mcorzEQ","e":"AQAB","d":"bMEEIVUarP7FJFFjR2m6seB9kbH6mHeTLHmIeE5stsEHUGPmefCF0Cw3N9EqnJnzM8JA_Rv99s7XGEJi2qAiPiHR7OH_2eu8-qE2h0hVOBs1ZSg7nV87rZGHfK39GtKx-wbEGpvRHI3dqSqU3NXjvK6tLhNtnNrOpIyfRwGTd75f8JaDLT8Nu1jhWD5ZRWua2vq_yklFsDuPhePh2fjZJ6eAqIx5I3gkpwKue8IuWd5PTwvY3CAxH3qlWy9aGWDxohS-zL0mWOXg82ZM8jE1fEP0Y_kFmvR8lfQU7HYyi6HAxKatYCuWaRf83OT5F3R7_adCTW7eZjSvcpKuhOsBFQ","p":"2dB5m17kN-H90rN7dNZ9fMf_t7FFV4LxuYA2MvwZYC8SFGaMmdgx20LGNBk4Dx9Mz4p8t06CY_9T_0Djk9F8URlrQJl7tzVN8Qq-FHcnjAY9zV_DINfbAWoQ4ZQT3I8syiMbAK8uOK7ehXbR9IbaikfJjJc5d9Vg2BNAACGRL68","q":"y2rfZzUlw2JScRhq9T0BzVuFVSOpad13qKzVDs13EmFoGqk31Zt1EnvaiwU7uGu2bGTTgQK65V90C5zgZW07yz1A7tSVGEDWgi3rYznMQzCUUHXRx7cX9i5G8OnBDKFznVvGXKSOFVq3AP2cOkr3xmF25ST--xwG2Ca0lTNB-T8","dp":"oXQyICRHoNDIuB1IvwObAxqxB7XEg6jRi0JpaoOKP8zEZxDY2dTyp-eoScgD0NnPsuuhpLLyXjNOTSAJUXHv56Gi6cCbfuNpQepHmZ31V4rs1sZMOpUmhrbbioqb6lrKxY8eHfS8m1GsKlw4Jzyq0-OAl9EkzRoC7kfeofo_x4s","dq":"WuV9wJuiLUWxOzJDESTauk4MLXhLCrBY-PmKFxw--eqm30sAVSYrUUAg7wA-qHERSixfyoVSyI43x7ypFQmTr4TGkDJUEUtfzzn_tg4stVVu4OlU_V5Wib4yGxMJHcDDbeyFnf42M1qe7gVlmzLGt1H0E_7NJZ5nfI0HIqiN8Xc","qi":"umm077LJtRZ8A19-xYSglcc32423DS-nHz5s3sW-EdX3wKJCG1A4zFpQ3gNTcBtIlBMIcAQYspJVE78cxQCoZ5D9Br2BJBcTLOS0tOvRzSyacXtiBDK0d8dZ_8rf1pfDg1Gt-isN9JaUgPITeT7b-WaBBLFz_f5mWH0vK97fnAQ"}`

	jwk, err := svc.GenerateJWKS()
	assert.NoError(t, err)
	assert.Equal(t, expectedJWK, jwk)
}

func TestGenerateJwksNoKeyAlgorithm(t *testing.T) {
	svc := New(nil)
	svc.Options.PrivateKey = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBclJNKzFQaWIwb2VGQ0dPbzVtTmpzTjh4T3JML242SXQwTXZ1TzN3UmRwQ2V3SmZzClY3SVJwWW5MNU5MenRXZnVoNm9rbGd2cTIwak00NDUxQXpIM2tuZGt4NFhqY090Zmg5WklSby9xQlhYeEdGOVgKSldXem1zMjFqWkJzOTVaL3p3VFpncEJVL21WaEhSejVuOStVR1NlRVpZZnU4WnlMeEZGQ3JvQlY1Uk5rTnJIQwowSWJOT1ZuOFJ1Snprczg5dnl2UzZES2xLY29IZmppVVgwTE1hNUZxcStuRSs5R3kzUURzcWRrYUtwR0JUbmFrCnNlWkIraEV2TFBGWDFUMFRHZDhzQTQ4SzEyWjVZaklaNjBPRHppTnB5dktuL24zWW9qWjFMZno5QzFSbDYwM3EKb2VuYzF3Y3hZYis1WGxYempVZDdKU0ZxZlRPSHR4L21jb3J6RVFJREFRQUJBb0lCQUd6QkJDRlZHcXoreFNSUgpZMGRwdXJIZ2ZaR3grcGgza3l4NWlIaE9iTGJCQjFCajVubndoZEFzTnpmUktweVo4elBDUVAwYi9mYk8xeGhDCll0cWdJajRoMGV6aC85bnJ2UHFoTm9kSVZUZ2JOV1VvTzUxZk82MlJoM3l0L1JyU3Nmc0d4QnFiMFJ5TjNha3EKbE56VjQ3eXVyUzRUYlp6YXpxU01uMGNCazNlK1gvQ1dneTAvRGJ0WTRWZytXVVZybXRyNnY4cEpSYkE3ajRYago0ZG40MlNlbmdLaU1lU040SktjQ3JudkNMbG5lVDA4TDJOd2dNUjk2cFZzdldobGc4YUlVdnN5OUpsamw0UE5tClRQSXhOWHhEOUdQNUJacjBmSlgwRk94Mk1vdWh3TVNtcldBcmxta1gvTnprK1JkMGUvMm5RazF1M21ZMHIzS1MKcm9UckFSVUNnWUVBMmRCNW0xN2tOK0g5MHJON2ROWjlmTWYvdDdGRlY0THh1WUEyTXZ3WllDOFNGR2FNbWRneAoyMExHTkJrNER4OU16NHA4dDA2Q1kvOVQvMERqazlGOFVSbHJRSmw3dHpWTjhRcStGSGNuakFZOXpWL0RJTmZiCkFXb1E0WlFUM0k4c3lpTWJBSzh1T0s3ZWhYYlI5SWJhaWtmSmpKYzVkOVZnMkJOQUFDR1JMNjhDZ1lFQXkycmYKWnpVbHcySlNjUmhxOVQwQnpWdUZWU09wYWQxM3FLelZEczEzRW1Gb0dxazMxWnQxRW52YWl3VTd1R3UyYkdUVApnUUs2NVY5MEM1emdaVzA3eXoxQTd0U1ZHRURXZ2kzcll6bk1RekNVVUhYUng3Y1g5aTVHOE9uQkRLRnpuVnZHClhLU09GVnEzQVAyY09rcjN4bUYyNVNUKyt4d0cyQ2EwbFROQitUOENnWUVBb1hReUlDUkhvTkRJdUIxSXZ3T2IKQXhxeEI3WEVnNmpSaTBKcGFvT0tQOHpFWnhEWTJkVHlwK2VvU2NnRDBOblBzdXVocExMeVhqTk9UU0FKVVhIdgo1NkdpNmNDYmZ1TnBRZXBIbVozMVY0cnMxc1pNT3BVbWhyYmJpb3FiNmxyS3hZOGVIZlM4bTFHc0tsdzRKenlxCjArT0FsOUVrelJvQzdrZmVvZm8veDRzQ2dZQmE1WDNBbTZJdFJiRTdNa01SSk5xNlRnd3RlRXNLc0ZqNCtZb1gKSEQ3NTZxYmZTd0JWSml0UlFDRHZBRDZvY1JGS0xGL0toVkxJampmSHZLa1ZDWk92aE1hUU1sUVJTMS9QT2YrMgpEaXkxVlc3ZzZWVDlYbGFKdmpJYkV3a2R3TU50N0lXZC9qWXpXcDd1QldXYk1zYTNVZlFUL3MwbG5tZDhqUWNpCnFJM3hkd0tCZ1FDNmFiVHZzc20xRm53RFgzN0ZoS0NWeHpmYmpiY05MNmNmUG16ZXhiNFIxZmZBb2tJYlVEak0KV2xEZUExTndHMGlVRXdod0JCaXlrbFVUdnh6RkFLaG5rUDBHdllFa0Z4TXM1TFMwNjlITkxKcHhlMklFTXJSMwp4MW4veXQvV2w4T0RVYTM2S3czMGxwU0E4aE41UHR2NVpvRUVzWFA5L21aWWZTOHIzdCtjQkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ=="
	svc.Options.Algorithm = ""

	expectedJWK := `{"kty":"RSA","kid":"ee8146d4b30a57d0053f39c80f4c3caa46461633","alg":"RS256","n":"rRM-1Pib0oeFCGOo5mNjsN8xOrL_n6It0MvuO3wRdpCewJfsV7IRpYnL5NLztWfuh6oklgvq20jM4451AzH3kndkx4XjcOtfh9ZIRo_qBXXxGF9XJWWzms21jZBs95Z_zwTZgpBU_mVhHRz5n9-UGSeEZYfu8ZyLxFFCroBV5RNkNrHC0IbNOVn8RuJzks89vyvS6DKlKcoHfjiUX0LMa5Fqq-nE-9Gy3QDsqdkaKpGBTnakseZB-hEvLPFX1T0TGd8sA48K12Z5YjIZ60ODziNpyvKn_n3YojZ1Lfz9C1Rl603qoenc1wcxYb-5XlXzjUd7JSFqfTOHtx_mcorzEQ","e":"AQAB","d":"bMEEIVUarP7FJFFjR2m6seB9kbH6mHeTLHmIeE5stsEHUGPmefCF0Cw3N9EqnJnzM8JA_Rv99s7XGEJi2qAiPiHR7OH_2eu8-qE2h0hVOBs1ZSg7nV87rZGHfK39GtKx-wbEGpvRHI3dqSqU3NXjvK6tLhNtnNrOpIyfRwGTd75f8JaDLT8Nu1jhWD5ZRWua2vq_yklFsDuPhePh2fjZJ6eAqIx5I3gkpwKue8IuWd5PTwvY3CAxH3qlWy9aGWDxohS-zL0mWOXg82ZM8jE1fEP0Y_kFmvR8lfQU7HYyi6HAxKatYCuWaRf83OT5F3R7_adCTW7eZjSvcpKuhOsBFQ","p":"2dB5m17kN-H90rN7dNZ9fMf_t7FFV4LxuYA2MvwZYC8SFGaMmdgx20LGNBk4Dx9Mz4p8t06CY_9T_0Djk9F8URlrQJl7tzVN8Qq-FHcnjAY9zV_DINfbAWoQ4ZQT3I8syiMbAK8uOK7ehXbR9IbaikfJjJc5d9Vg2BNAACGRL68","q":"y2rfZzUlw2JScRhq9T0BzVuFVSOpad13qKzVDs13EmFoGqk31Zt1EnvaiwU7uGu2bGTTgQK65V90C5zgZW07yz1A7tSVGEDWgi3rYznMQzCUUHXRx7cX9i5G8OnBDKFznVvGXKSOFVq3AP2cOkr3xmF25ST--xwG2Ca0lTNB-T8","dp":"oXQyICRHoNDIuB1IvwObAxqxB7XEg6jRi0JpaoOKP8zEZxDY2dTyp-eoScgD0NnPsuuhpLLyXjNOTSAJUXHv56Gi6cCbfuNpQepHmZ31V4rs1sZMOpUmhrbbioqb6lrKxY8eHfS8m1GsKlw4Jzyq0-OAl9EkzRoC7kfeofo_x4s","dq":"WuV9wJuiLUWxOzJDESTauk4MLXhLCrBY-PmKFxw--eqm30sAVSYrUUAg7wA-qHERSixfyoVSyI43x7ypFQmTr4TGkDJUEUtfzzn_tg4stVVu4OlU_V5Wib4yGxMJHcDDbeyFnf42M1qe7gVlmzLGt1H0E_7NJZ5nfI0HIqiN8Xc","qi":"umm077LJtRZ8A19-xYSglcc32423DS-nHz5s3sW-EdX3wKJCG1A4zFpQ3gNTcBtIlBMIcAQYspJVE78cxQCoZ5D9Br2BJBcTLOS0tOvRzSyacXtiBDK0d8dZ_8rf1pfDg1Gt-isN9JaUgPITeT7b-WaBBLFz_f5mWH0vK97fnAQ"}`

	jwk, err := svc.GenerateJWKS()
	assert.NoError(t, err)
	assert.Equal(t, expectedJWK, jwk)
}

func TestGenerateJWKSEmptyPrivateKey(t *testing.T) {
	svc := New(nil)
	svc.Options.PrivateKey = ""

	_, err := svc.GenerateJWKS()
	assert.EqualError(t, err, "private key cannot be empty")
}

func TestGenerateJWKSInvalidPrivateKey(t *testing.T) {
	svc := New(nil)
	svc.Options.PrivateKey = "invalidPrivateKey"

	_, err := svc.GenerateJWKS()
	assert.Error(t, err)
}

func TestGenerateJWKSRS384Algorithm(t *testing.T) {
	svc := New(nil)
	svc.Options.PrivateKey = "cGFzc3dvcmQ="
	svc.Options.Algorithm = JwtSigningAlgorithmRS384

	_, err := svc.GenerateJWKS()
	assert.Error(t, err, "private key cannot be empty")
}

func TestVerifyJwtWithHS256Algorithm(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.WithSecret("secret")
	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	tokenStr, err := svc.Sign(claims)
	assert.NoError(t, err)

	token, err := svc.Parse(tokenStr)
	assert.NoError(t, err)

	verifiedToken, err := token.Valid()
	assert.NoError(t, err)

	assert.True(t, verifiedToken)
}

func TestVerifyJwtWithHS256AlgorithmWithNoSecret(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.WithSecret("secret")
	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	tokenStr, err := svc.Sign(claims)
	assert.NoError(t, err)

	svc.Options.Secret = ""
	_, err = svc.Parse(tokenStr)
	assert.Errorf(t, err, "secret cannot be empty")
}

func TestVerifyJwtWithRS256Algorithm(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.WithPrivateKey("LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBclJNKzFQaWIwb2VGQ0dPbzVtTmpzTjh4T3JML242SXQwTXZ1TzN3UmRwQ2V3SmZzClY3SVJwWW5MNU5MenRXZnVoNm9rbGd2cTIwak00NDUxQXpIM2tuZGt4NFhqY090Zmg5WklSby9xQlhYeEdGOVgKSldXem1zMjFqWkJzOTVaL3p3VFpncEJVL21WaEhSejVuOStVR1NlRVpZZnU4WnlMeEZGQ3JvQlY1Uk5rTnJIQwowSWJOT1ZuOFJ1Snprczg5dnl2UzZES2xLY29IZmppVVgwTE1hNUZxcStuRSs5R3kzUURzcWRrYUtwR0JUbmFrCnNlWkIraEV2TFBGWDFUMFRHZDhzQTQ4SzEyWjVZaklaNjBPRHppTnB5dktuL24zWW9qWjFMZno5QzFSbDYwM3EKb2VuYzF3Y3hZYis1WGxYempVZDdKU0ZxZlRPSHR4L21jb3J6RVFJREFRQUJBb0lCQUd6QkJDRlZHcXoreFNSUgpZMGRwdXJIZ2ZaR3grcGgza3l4NWlIaE9iTGJCQjFCajVubndoZEFzTnpmUktweVo4elBDUVAwYi9mYk8xeGhDCll0cWdJajRoMGV6aC85bnJ2UHFoTm9kSVZUZ2JOV1VvTzUxZk82MlJoM3l0L1JyU3Nmc0d4QnFiMFJ5TjNha3EKbE56VjQ3eXVyUzRUYlp6YXpxU01uMGNCazNlK1gvQ1dneTAvRGJ0WTRWZytXVVZybXRyNnY4cEpSYkE3ajRYago0ZG40MlNlbmdLaU1lU040SktjQ3JudkNMbG5lVDA4TDJOd2dNUjk2cFZzdldobGc4YUlVdnN5OUpsamw0UE5tClRQSXhOWHhEOUdQNUJacjBmSlgwRk94Mk1vdWh3TVNtcldBcmxta1gvTnprK1JkMGUvMm5RazF1M21ZMHIzS1MKcm9UckFSVUNnWUVBMmRCNW0xN2tOK0g5MHJON2ROWjlmTWYvdDdGRlY0THh1WUEyTXZ3WllDOFNGR2FNbWRneAoyMExHTkJrNER4OU16NHA4dDA2Q1kvOVQvMERqazlGOFVSbHJRSmw3dHpWTjhRcStGSGNuakFZOXpWL0RJTmZiCkFXb1E0WlFUM0k4c3lpTWJBSzh1T0s3ZWhYYlI5SWJhaWtmSmpKYzVkOVZnMkJOQUFDR1JMNjhDZ1lFQXkycmYKWnpVbHcySlNjUmhxOVQwQnpWdUZWU09wYWQxM3FLelZEczEzRW1Gb0dxazMxWnQxRW52YWl3VTd1R3UyYkdUVApnUUs2NVY5MEM1emdaVzA3eXoxQTd0U1ZHRURXZ2kzcll6bk1RekNVVUhYUng3Y1g5aTVHOE9uQkRLRnpuVnZHClhLU09GVnEzQVAyY09rcjN4bUYyNVNUKyt4d0cyQ2EwbFROQitUOENnWUVBb1hReUlDUkhvTkRJdUIxSXZ3T2IKQXhxeEI3WEVnNmpSaTBKcGFvT0tQOHpFWnhEWTJkVHlwK2VvU2NnRDBOblBzdXVocExMeVhqTk9UU0FKVVhIdgo1NkdpNmNDYmZ1TnBRZXBIbVozMVY0cnMxc1pNT3BVbWhyYmJpb3FiNmxyS3hZOGVIZlM4bTFHc0tsdzRKenlxCjArT0FsOUVrelJvQzdrZmVvZm8veDRzQ2dZQmE1WDNBbTZJdFJiRTdNa01SSk5xNlRnd3RlRXNLc0ZqNCtZb1gKSEQ3NTZxYmZTd0JWSml0UlFDRHZBRDZvY1JGS0xGL0toVkxJampmSHZLa1ZDWk92aE1hUU1sUVJTMS9QT2YrMgpEaXkxVlc3ZzZWVDlYbGFKdmpJYkV3a2R3TU50N0lXZC9qWXpXcDd1QldXYk1zYTNVZlFUL3MwbG5tZDhqUWNpCnFJM3hkd0tCZ1FDNmFiVHZzc20xRm53RFgzN0ZoS0NWeHpmYmpiY05MNmNmUG16ZXhiNFIxZmZBb2tJYlVEak0KV2xEZUExTndHMGlVRXdod0JCaXlrbFVUdnh6RkFLaG5rUDBHdllFa0Z4TXM1TFMwNjlITkxKcHhlMklFTXJSMwp4MW4veXQvV2w4T0RVYTM2S3czMGxwU0E4aE41UHR2NVpvRUVzWFA5L21aWWZTOHIzdCtjQkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ==")
	svc.Options.Algorithm = JwtSigningAlgorithmRS256

	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	tokenStr, err := svc.Sign(claims)
	assert.NoError(t, err)

	token, err := svc.Parse(tokenStr)
	assert.NoError(t, err)

	verifiedToken, err := token.Valid()
	assert.NoError(t, err)

	assert.True(t, verifiedToken)
}

func TestVerifyJwtWithRS256AlgorithmWithNoPrivateKey(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.Options.Algorithm = JwtSigningAlgorithmRS256
	svc.WithPrivateKey("LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBclJNKzFQaWIwb2VGQ0dPbzVtTmpzTjh4T3JML242SXQwTXZ1TzN3UmRwQ2V3SmZzClY3SVJwWW5MNU5MenRXZnVoNm9rbGd2cTIwak00NDUxQXpIM2tuZGt4NFhqY090Zmg5WklSby9xQlhYeEdGOVgKSldXem1zMjFqWkJzOTVaL3p3VFpncEJVL21WaEhSejVuOStVR1NlRVpZZnU4WnlMeEZGQ3JvQlY1Uk5rTnJIQwowSWJOT1ZuOFJ1Snprczg5dnl2UzZES2xLY29IZmppVVgwTE1hNUZxcStuRSs5R3kzUURzcWRrYUtwR0JUbmFrCnNlWkIraEV2TFBGWDFUMFRHZDhzQTQ4SzEyWjVZaklaNjBPRHppTnB5dktuL24zWW9qWjFMZno5QzFSbDYwM3EKb2VuYzF3Y3hZYis1WGxYempVZDdKU0ZxZlRPSHR4L21jb3J6RVFJREFRQUJBb0lCQUd6QkJDRlZHcXoreFNSUgpZMGRwdXJIZ2ZaR3grcGgza3l4NWlIaE9iTGJCQjFCajVubndoZEFzTnpmUktweVo4elBDUVAwYi9mYk8xeGhDCll0cWdJajRoMGV6aC85bnJ2UHFoTm9kSVZUZ2JOV1VvTzUxZk82MlJoM3l0L1JyU3Nmc0d4QnFiMFJ5TjNha3EKbE56VjQ3eXVyUzRUYlp6YXpxU01uMGNCazNlK1gvQ1dneTAvRGJ0WTRWZytXVVZybXRyNnY4cEpSYkE3ajRYago0ZG40MlNlbmdLaU1lU040SktjQ3JudkNMbG5lVDA4TDJOd2dNUjk2cFZzdldobGc4YUlVdnN5OUpsamw0UE5tClRQSXhOWHhEOUdQNUJacjBmSlgwRk94Mk1vdWh3TVNtcldBcmxta1gvTnprK1JkMGUvMm5RazF1M21ZMHIzS1MKcm9UckFSVUNnWUVBMmRCNW0xN2tOK0g5MHJON2ROWjlmTWYvdDdGRlY0THh1WUEyTXZ3WllDOFNGR2FNbWRneAoyMExHTkJrNER4OU16NHA4dDA2Q1kvOVQvMERqazlGOFVSbHJRSmw3dHpWTjhRcStGSGNuakFZOXpWL0RJTmZiCkFXb1E0WlFUM0k4c3lpTWJBSzh1T0s3ZWhYYlI5SWJhaWtmSmpKYzVkOVZnMkJOQUFDR1JMNjhDZ1lFQXkycmYKWnpVbHcySlNjUmhxOVQwQnpWdUZWU09wYWQxM3FLelZEczEzRW1Gb0dxazMxWnQxRW52YWl3VTd1R3UyYkdUVApnUUs2NVY5MEM1emdaVzA3eXoxQTd0U1ZHRURXZ2kzcll6bk1RekNVVUhYUng3Y1g5aTVHOE9uQkRLRnpuVnZHClhLU09GVnEzQVAyY09rcjN4bUYyNVNUKyt4d0cyQ2EwbFROQitUOENnWUVBb1hReUlDUkhvTkRJdUIxSXZ3T2IKQXhxeEI3WEVnNmpSaTBKcGFvT0tQOHpFWnhEWTJkVHlwK2VvU2NnRDBOblBzdXVocExMeVhqTk9UU0FKVVhIdgo1NkdpNmNDYmZ1TnBRZXBIbVozMVY0cnMxc1pNT3BVbWhyYmJpb3FiNmxyS3hZOGVIZlM4bTFHc0tsdzRKenlxCjArT0FsOUVrelJvQzdrZmVvZm8veDRzQ2dZQmE1WDNBbTZJdFJiRTdNa01SSk5xNlRnd3RlRXNLc0ZqNCtZb1gKSEQ3NTZxYmZTd0JWSml0UlFDRHZBRDZvY1JGS0xGL0toVkxJampmSHZLa1ZDWk92aE1hUU1sUVJTMS9QT2YrMgpEaXkxVlc3ZzZWVDlYbGFKdmpJYkV3a2R3TU50N0lXZC9qWXpXcDd1QldXYk1zYTNVZlFUL3MwbG5tZDhqUWNpCnFJM3hkd0tCZ1FDNmFiVHZzc20xRm53RFgzN0ZoS0NWeHpmYmpiY05MNmNmUG16ZXhiNFIxZmZBb2tJYlVEak0KV2xEZUExTndHMGlVRXdod0JCaXlrbFVUdnh6RkFLaG5rUDBHdllFa0Z4TXM1TFMwNjlITkxKcHhlMklFTXJSMwp4MW4veXQvV2w4T0RVYTM2S3czMGxwU0E4aE41UHR2NVpvRUVzWFA5L21aWWZTOHIzdCtjQkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ==")

	claims := map[string]interface{}{
		"email":  "test@example.com",
		"roles":  []string{"admin", "user"},
		"claims": []string{"claim1", "claim2"},
	}

	tokenStr, err := svc.Sign(claims)
	assert.NoError(t, err)

	svc.Options.PrivateKey = ""
	_, err = svc.Parse(tokenStr)

	assert.Errorf(t, err, "")
}

func TestVerifyJJwtWithNoRolesAndClaims(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)
	svc.Options.Algorithm = JwtSigningAlgorithmRS256
	svc.WithPrivateKey("LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBclJNKzFQaWIwb2VGQ0dPbzVtTmpzTjh4T3JML242SXQwTXZ1TzN3UmRwQ2V3SmZzClY3SVJwWW5MNU5MenRXZnVoNm9rbGd2cTIwak00NDUxQXpIM2tuZGt4NFhqY090Zmg5WklSby9xQlhYeEdGOVgKSldXem1zMjFqWkJzOTVaL3p3VFpncEJVL21WaEhSejVuOStVR1NlRVpZZnU4WnlMeEZGQ3JvQlY1Uk5rTnJIQwowSWJOT1ZuOFJ1Snprczg5dnl2UzZES2xLY29IZmppVVgwTE1hNUZxcStuRSs5R3kzUURzcWRrYUtwR0JUbmFrCnNlWkIraEV2TFBGWDFUMFRHZDhzQTQ4SzEyWjVZaklaNjBPRHppTnB5dktuL24zWW9qWjFMZno5QzFSbDYwM3EKb2VuYzF3Y3hZYis1WGxYempVZDdKU0ZxZlRPSHR4L21jb3J6RVFJREFRQUJBb0lCQUd6QkJDRlZHcXoreFNSUgpZMGRwdXJIZ2ZaR3grcGgza3l4NWlIaE9iTGJCQjFCajVubndoZEFzTnpmUktweVo4elBDUVAwYi9mYk8xeGhDCll0cWdJajRoMGV6aC85bnJ2UHFoTm9kSVZUZ2JOV1VvTzUxZk82MlJoM3l0L1JyU3Nmc0d4QnFiMFJ5TjNha3EKbE56VjQ3eXVyUzRUYlp6YXpxU01uMGNCazNlK1gvQ1dneTAvRGJ0WTRWZytXVVZybXRyNnY4cEpSYkE3ajRYago0ZG40MlNlbmdLaU1lU040SktjQ3JudkNMbG5lVDA4TDJOd2dNUjk2cFZzdldobGc4YUlVdnN5OUpsamw0UE5tClRQSXhOWHhEOUdQNUJacjBmSlgwRk94Mk1vdWh3TVNtcldBcmxta1gvTnprK1JkMGUvMm5RazF1M21ZMHIzS1MKcm9UckFSVUNnWUVBMmRCNW0xN2tOK0g5MHJON2ROWjlmTWYvdDdGRlY0THh1WUEyTXZ3WllDOFNGR2FNbWRneAoyMExHTkJrNER4OU16NHA4dDA2Q1kvOVQvMERqazlGOFVSbHJRSmw3dHpWTjhRcStGSGNuakFZOXpWL0RJTmZiCkFXb1E0WlFUM0k4c3lpTWJBSzh1T0s3ZWhYYlI5SWJhaWtmSmpKYzVkOVZnMkJOQUFDR1JMNjhDZ1lFQXkycmYKWnpVbHcySlNjUmhxOVQwQnpWdUZWU09wYWQxM3FLelZEczEzRW1Gb0dxazMxWnQxRW52YWl3VTd1R3UyYkdUVApnUUs2NVY5MEM1emdaVzA3eXoxQTd0U1ZHRURXZ2kzcll6bk1RekNVVUhYUng3Y1g5aTVHOE9uQkRLRnpuVnZHClhLU09GVnEzQVAyY09rcjN4bUYyNVNUKyt4d0cyQ2EwbFROQitUOENnWUVBb1hReUlDUkhvTkRJdUIxSXZ3T2IKQXhxeEI3WEVnNmpSaTBKcGFvT0tQOHpFWnhEWTJkVHlwK2VvU2NnRDBOblBzdXVocExMeVhqTk9UU0FKVVhIdgo1NkdpNmNDYmZ1TnBRZXBIbVozMVY0cnMxc1pNT3BVbWhyYmJpb3FiNmxyS3hZOGVIZlM4bTFHc0tsdzRKenlxCjArT0FsOUVrelJvQzdrZmVvZm8veDRzQ2dZQmE1WDNBbTZJdFJiRTdNa01SSk5xNlRnd3RlRXNLc0ZqNCtZb1gKSEQ3NTZxYmZTd0JWSml0UlFDRHZBRDZvY1JGS0xGL0toVkxJampmSHZLa1ZDWk92aE1hUU1sUVJTMS9QT2YrMgpEaXkxVlc3ZzZWVDlYbGFKdmpJYkV3a2R3TU50N0lXZC9qWXpXcDd1QldXYk1zYTNVZlFUL3MwbG5tZDhqUWNpCnFJM3hkd0tCZ1FDNmFiVHZzc20xRm53RFgzN0ZoS0NWeHpmYmpiY05MNmNmUG16ZXhiNFIxZmZBb2tJYlVEak0KV2xEZUExTndHMGlVRXdod0JCaXlrbFVUdnh6RkFLaG5rUDBHdllFa0Z4TXM1TFMwNjlITkxKcHhlMklFTXJSMwp4MW4veXQvV2w4T0RVYTM2S3czMGxwU0E4aE41UHR2NVpvRUVzWFA5L21aWWZTOHIzdCtjQkE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQ==")

	claims := map[string]interface{}{
		"email": "test@example.com",
	}

	tokenStr, err := svc.Sign(claims)
	assert.NoError(t, err)

	_, err = svc.Parse(tokenStr)
	assert.NoError(t, err)
}

func TestJwtService_processEnvironmentVariables(t *testing.T) {
	ctx := basecontext.NewBaseContext()
	svc := New(ctx)

	t.Run("SetAlgorithm", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.JWT_SIGN_ALGORITHM_ENV_VAR, "HS256")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.NoError(t, err)

		assert.Equal(t, JwtSigningAlgorithmHS256, svc.Options.Algorithm)
	})

	t.Run("SetInvalidAlgorithm", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.JWT_SIGN_ALGORITHM_ENV_VAR, "invalid")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.Error(t, err)
		assert.Equal(t, errors.New("invalid signing algorithm"), err)
	})

	t.Run("SetSecret", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.JWT_HMACS_SECRET_ENV_VAR, "secret")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.NoError(t, err)

		assert.Equal(t, "secret", svc.Options.Secret)
	})

	t.Run("SetPrivateKey", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.JWT_PRIVATE_KEY_ENV_VAR, "private_key")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.NoError(t, err)

		assert.Equal(t, "private_key", svc.Options.PrivateKey)
	})

	t.Run("SetTokenDuration", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.JWT_DURATION_ENV_VAR, "60m")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.NoError(t, err)

		assert.Equal(t, "60m", svc.Options.TokenDuration)
	})

	t.Run("InvalidTokenDuration", func(t *testing.T) {
		os.Clearenv()
		err := os.Setenv(constants.JWT_DURATION_ENV_VAR, "invalid")
		assert.NoError(t, err)

		err = svc.processEnvironmentVariables()
		assert.Error(t, err)
	})
}
