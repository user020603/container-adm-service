package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPasswordAndCheckPasswordHash(t *testing.T) {
	password := "secret123A@"

	hash, err := HashPassword(password)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	ok := CheckPasswordHash(password, hash)
	assert.True(t, ok, "Password should match the hash")

	ok = CheckPasswordHash("wrongpassword", hash)
	assert.False(t, ok, "Wrong password should not match the hash")
}