package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("my-password")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, "my-password", hash)
}

func TestComparePassword(t *testing.T) {
	hash, _ := HashPassword("my-password")

	t.Run("correct password", func(t *testing.T) {
		err := ComparePassword(hash, "my-password")
		assert.NoError(t, err)
	})

	t.Run("wrong password", func(t *testing.T) {
		err := ComparePassword(hash, "wrong-password")
		assert.Error(t, err)
	})
}
