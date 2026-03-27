package manager

import (
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestManager() *JWTManager {
	return NewJWTManager("test-secret-key", 1*time.Hour)
}

func TestNewJWTManager(t *testing.T) {
	m := NewJWTManager("secret", 2*time.Hour)
	assert.NotNil(t, m)
}

func TestGenerate_Success(t *testing.T) {
	m := newTestManager()
	token, err := m.Generate("user-123", "johndoe")
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestVerify_ValidToken(t *testing.T) {
	m := newTestManager()
	token, err := m.Generate("user-123", "johndoe")
	require.NoError(t, err)

	claims, err := m.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", claims.UserID)
	assert.Equal(t, "johndoe", claims.Username)
	assert.Equal(t, "user-123", claims.Subject)
}

func TestVerify_InvalidToken(t *testing.T) {
	m := newTestManager()
	claims, err := m.Verify("invalid-token")
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestVerify_WrongSecret(t *testing.T) {
	m1 := NewJWTManager("secret-1", 1*time.Hour)
	m2 := NewJWTManager("secret-2", 1*time.Hour)

	token, err := m1.Generate("user-123", "johndoe")
	require.NoError(t, err)

	claims, err := m2.Verify(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestVerify_ExpiredToken(t *testing.T) {
	m := NewJWTManager("secret", -1*time.Hour)
	token, err := m.Generate("user-123", "johndoe")
	require.NoError(t, err)

	claims, err := m.Verify(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestGenerate_DifferentUsers_DifferentTokens(t *testing.T) {
	m := newTestManager()
	token1, _ := m.Generate("user-1", "alice")
	token2, _ := m.Generate("user-2", "bob")
	assert.NotEqual(t, token1, token2)
}

func TestVerify_EmptyToken(t *testing.T) {
	m := newTestManager()
	claims, err := m.Verify("")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestVerify_MalformedToken(t *testing.T) {
	m := newTestManager()
	claims, err := m.Verify("a.b.c")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestVerify_TokenWithNoneAlgorithm(t *testing.T) {
	// Create token with "none" signing method
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"user_id":  "user-123",
		"username": "johndoe",
		"exp":      time.Now().Add(1 * time.Hour).Unix(),
	})
	tokenStr, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	m := newTestManager()
	claims, err := m.Verify(tokenStr)
	assert.Error(t, err)
	assert.Nil(t, claims)
}
