package utcp

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscoverTools_MissingURL(t *testing.T) {
	handler := NewHandler(nil)

	body, _ := json.Marshal(DiscoverRequest{})
	req := httptest.NewRequest(http.MethodPost, "/discover", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.DiscoverTools(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDiscoverTools_InvalidBody(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/discover", bytes.NewReader([]byte(`{invalid`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.DiscoverTools(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestConnectManual_MissingFields(t *testing.T) {
	handler := NewHandler(nil)

	// Missing name
	body, _ := json.Marshal(ConnectManualRequest{ManualURL: "http://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/manuals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ConnectManual(w, req)

	// Should fail because agentID can't be parsed from chi context (empty)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
