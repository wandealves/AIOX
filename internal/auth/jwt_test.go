package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	mgr := NewJWTManager("access-secret-32-chars-long!!!!!", "refresh-secret-32-chars-long!!!!", 15*time.Minute, 7*24*time.Hour)

	t.Run("generate and validate access token", func(t *testing.T) {
		pair, tokenID, err := mgr.GenerateTokenPair("user-123", "test@example.com")
		require.NoError(t, err)
		assert.NotEmpty(t, pair.AccessToken)
		assert.NotEmpty(t, pair.RefreshToken)
		assert.NotEmpty(t, tokenID)
		assert.Equal(t, int64(900), pair.ExpiresIn)

		claims, err := mgr.ValidateAccessToken(pair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, "user-123", claims.UserID)
		assert.Equal(t, "test@example.com", claims.Email)
	})

	t.Run("generate and validate refresh token", func(t *testing.T) {
		pair, _, err := mgr.GenerateTokenPair("user-456", "user@example.com")
		require.NoError(t, err)

		claims, err := mgr.ValidateRefreshToken(pair.RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, "user-456", claims.UserID)
		assert.NotEmpty(t, claims.TokenID)
	})

	t.Run("invalid token fails validation", func(t *testing.T) {
		_, err := mgr.ValidateAccessToken("invalid-token")
		assert.Error(t, err)
	})

	t.Run("access token cant validate as refresh", func(t *testing.T) {
		pair, _, _ := mgr.GenerateTokenPair("user-789", "x@x.com")
		_, err := mgr.ValidateRefreshToken(pair.AccessToken)
		assert.Error(t, err)
	})

	t.Run("expired token fails", func(t *testing.T) {
		shortMgr := NewJWTManager("access-secret-32-chars-long!!!!!", "refresh-secret-32-chars-long!!!!", -1*time.Second, -1*time.Second)
		pair, _, err := shortMgr.GenerateTokenPair("user-exp", "exp@test.com")
		require.NoError(t, err)

		_, err = shortMgr.ValidateAccessToken(pair.AccessToken)
		assert.Error(t, err)
	})
}
