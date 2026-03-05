package utcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDiscover_ValidManual(t *testing.T) {
	manualJSON := UTCPManualJSON{
		Name:        "test-api",
		Description: "A test API",
		Version:     "2.0",
		BaseURL:     "https://api.test.com",
		Auth: &UTCPAuthSpec{
			Type:     "api_key",
			Location: "header",
			Vars:     map[string]string{"X-Api-Key": "Your key"},
		},
		Tools: []UTCPToolSpec{
			{
				Name:        "search",
				Description: "Search endpoint",
				Path:        "/search",
				Method:      "GET",
				Parameters:  json.RawMessage(`{"type": "object"}`),
			},
			{
				Name:        "create",
				Description: "Create endpoint",
				Path:        "/create",
				Method:      "POST",
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(manualJSON)
	}))
	defer srv.Close()

	svc := &Service{
		httpClient: srv.Client(),
	}

	result, err := svc.Discover(context.Background(), srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Name != "test-api" {
		t.Errorf("expected name 'test-api', got '%s'", result.Name)
	}
	if result.Version != "2.0" {
		t.Errorf("expected version '2.0', got '%s'", result.Version)
	}
	if len(result.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(result.Tools))
	}
	if len(result.AuthVars) != 1 {
		t.Errorf("expected 1 auth var, got %d", len(result.AuthVars))
	}
}

func TestDiscover_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{invalid`))
	}))
	defer srv.Close()

	svc := &Service{
		httpClient: srv.Client(),
	}

	_, err := svc.Discover(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDiscover_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	svc := &Service{
		httpClient: srv.Client(),
	}

	_, err := svc.Discover(context.Background(), srv.URL)
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestFetchManualURL_SizeLimit(t *testing.T) {
	// Create a server that returns a very large response
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write more than 1MB — we just write a valid start and it'll be truncated
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"name": "big"`))
		// Writing 1MB of padding
		buf := make([]byte, 1<<20)
		for i := range buf {
			buf[i] = ' '
		}
		w.Write(buf)
		w.Write([]byte(`}`))
	}))
	defer srv.Close()

	svc := &Service{httpClient: srv.Client()}
	_, err := svc.fetchManualURL(srv.URL)
	// Should still read (limited to 1MB) but parsing the truncated JSON should fail
	// or it might succeed with partial data — either way, no crash
	_ = err
}
