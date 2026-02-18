//go:build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
	env := SetupTestEnv(t)

	resp := DoRequest(t, env, "GET", "/health", nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	result := ParseResponse(t, resp)
	data := result["data"].(map[string]any)
	assert.Equal(t, "healthy", data["status"])
}

func TestRegister(t *testing.T) {
	env := SetupTestEnv(t)

	t.Run("successful registration", func(t *testing.T) {
		result := RegisterUser(t, env, "test-reg@example.com", "password123")
		data := result["data"].(map[string]any)

		assert.NotEmpty(t, data["access_token"])
		assert.NotEmpty(t, data["refresh_token"])
		assert.NotZero(t, data["expires_in"])
	})

	t.Run("duplicate email", func(t *testing.T) {
		RegisterUser(t, env, "dupe@example.com", "password123")

		body := map[string]string{"email": "dupe@example.com", "password": "password123"}
		resp := DoRequest(t, env, "POST", "/api/v1/auth/register", body, "")
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("invalid email", func(t *testing.T) {
		body := map[string]string{"email": "not-an-email", "password": "password123"}
		resp := DoRequest(t, env, "POST", "/api/v1/auth/register", body, "")
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("short password", func(t *testing.T) {
		body := map[string]string{"email": "short@example.com", "password": "short"}
		resp := DoRequest(t, env, "POST", "/api/v1/auth/register", body, "")
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestLogin(t *testing.T) {
	env := SetupTestEnv(t)

	RegisterUser(t, env, "login@example.com", "password123")

	t.Run("successful login", func(t *testing.T) {
		body := map[string]string{"email": "login@example.com", "password": "password123"}
		resp := DoRequest(t, env, "POST", "/api/v1/auth/login", body, "")
		require.Equal(t, http.StatusOK, resp.StatusCode)

		result := ParseResponse(t, resp)
		data := result["data"].(map[string]any)
		assert.NotEmpty(t, data["access_token"])
		assert.NotEmpty(t, data["refresh_token"])
	})

	t.Run("wrong password", func(t *testing.T) {
		body := map[string]string{"email": "login@example.com", "password": "wrongpass"}
		resp := DoRequest(t, env, "POST", "/api/v1/auth/login", body, "")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("non-existent user", func(t *testing.T) {
		body := map[string]string{"email": "nobody@example.com", "password": "password123"}
		resp := DoRequest(t, env, "POST", "/api/v1/auth/login", body, "")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestRefreshToken(t *testing.T) {
	env := SetupTestEnv(t)

	result := RegisterUser(t, env, "refresh@example.com", "password123")
	data := result["data"].(map[string]any)
	refreshToken := data["refresh_token"].(string)

	t.Run("successful refresh", func(t *testing.T) {
		body := map[string]string{"refresh_token": refreshToken}
		resp := DoRequest(t, env, "POST", "/api/v1/auth/refresh", body, "")
		require.Equal(t, http.StatusOK, resp.StatusCode)

		newResult := ParseResponse(t, resp)
		newData := newResult["data"].(map[string]any)
		assert.NotEmpty(t, newData["access_token"])
		assert.NotEmpty(t, newData["refresh_token"])
	})

	t.Run("reuse old refresh token fails", func(t *testing.T) {
		body := map[string]string{"refresh_token": refreshToken}
		resp := DoRequest(t, env, "POST", "/api/v1/auth/refresh", body, "")
		// Should fail because old token was rotated
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		body := map[string]string{"refresh_token": "invalid-token"}
		resp := DoRequest(t, env, "POST", "/api/v1/auth/refresh", body, "")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestLogout(t *testing.T) {
	env := SetupTestEnv(t)

	RegisterUser(t, env, "logout@example.com", "password123")
	token := LoginUser(t, env, "logout@example.com", "password123")

	t.Run("successful logout", func(t *testing.T) {
		resp := DoRequest(t, env, "POST", "/api/v1/auth/logout", nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("logout without token", func(t *testing.T) {
		resp := DoRequest(t, env, "POST", "/api/v1/auth/logout", nil, "")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
