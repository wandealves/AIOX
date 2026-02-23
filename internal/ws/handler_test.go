package ws

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aiox-platform/aiox/internal/auth"
)

func TestUpgrade_MissingToken(t *testing.T) {
	hub := NewHub()
	jwtMgr := auth.NewJWTManager("test-secret-12345678901234567890", "ref-secret-12345678901234567890", 15*time.Minute, 24*time.Hour)
	h := NewHandler(hub, jwtMgr, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws", nil)
	w := httptest.NewRecorder()

	h.Upgrade(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing token")
}

func TestUpgrade_InvalidToken(t *testing.T) {
	hub := NewHub()
	jwtMgr := auth.NewJWTManager("test-secret-12345678901234567890", "ref-secret-12345678901234567890", 15*time.Minute, 24*time.Hour)
	h := NewHandler(hub, jwtMgr, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ws?token=invalid-token", nil)
	w := httptest.NewRecorder()

	h.Upgrade(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid token")
}
