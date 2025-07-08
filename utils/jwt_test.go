package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateAndParseJWT(t *testing.T) {
	userID := uint(42)
	username := "testuser"
	role := "admin"
	expiry := time.Minute * 5

	scopes := []string{}
	tokenStr, err := GenerateJWT(userID, username, role, scopes, expiry)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr, "Token should not be empty")

	claims, err := ParseJWT(tokenStr)
	assert.NoError(t, err)
	assert.NotNil(t, claims, "Claims should not be nil")
	assert.Equal(t, userID, claims.UserID, "UserID should match")
	assert.Equal(t, username, claims.Username, "Username should match")
	assert.Equal(t, role, claims.Role, "Role should match")
}

func TestParseJWT_InvalidToken(t *testing.T) {
	invalidToken := "invalid.token.string"

	claims, err := ParseJWT(invalidToken)
	assert.Error(t, err, "Parsing an invalid token should return an error")
	assert.Nil(t, claims, "Claims should be nil for an invalid token")
}