package utcp

import (
	"encoding/json"
	"testing"
)

func TestParseManual_RealUTCPFormat(t *testing.T) {
	data := []byte(`{
		"manual_version": "1.0.0",
		"utcp_version": "1.0.1",
		"name": "UtcpApi",
		"description": "Product management API",
		"auth": {
			"auth_type": "bearer",
			"location": "header",
			"var_name": "Authorization",
			"prefix": "Bearer"
		},
		"tools": [
			{
				"name": "register",
				"description": "Register a user",
				"inputs": {"type": "object", "properties": {"username": {"type": "string"}}},
				"tool_provider": {
					"provider_type": "http",
					"http_method": "POST",
					"url": "http://localhost:5227/auth/register",
					"content_type": "application/json"
				}
			},
			{
				"name": "list_products",
				"description": "List products",
				"inputs": {"type": "object", "properties": {}},
				"tool_provider": {
					"provider_type": "http",
					"http_method": "GET",
					"url": "http://localhost:5227/products",
					"auth": {"auth_type": "bearer", "var_name": "Authorization", "location": "header"}
				}
			}
		]
	}`)

	manual, err := ParseManual(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if manual.Name != "UtcpApi" {
		t.Errorf("expected name 'UtcpApi', got '%s'", manual.Name)
	}
	if len(manual.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(manual.Tools))
	}

	// Check tool_provider URL extraction
	if url := manual.Tools[0].GetEndpointURL(""); url != "http://localhost:5227/auth/register" {
		t.Errorf("expected register URL, got '%s'", url)
	}
	if method := manual.Tools[0].GetHTTPMethod(); method != "POST" {
		t.Errorf("expected POST, got '%s'", method)
	}
	if url := manual.Tools[1].GetEndpointURL(""); url != "http://localhost:5227/products" {
		t.Errorf("expected products URL, got '%s'", url)
	}

	// Per-tool auth
	if manual.Tools[1].GetToolAuth() == nil {
		t.Error("expected per-tool auth on list_products")
	}
}

func TestParseManual_SimplifiedFormat(t *testing.T) {
	data := []byte(`{
		"name": "weather-api",
		"base_url": "https://api.weather.com",
		"auth": {"type": "api_key", "location": "header", "vars": {"X-Api-Key": "key"}},
		"tools": [
			{"name": "get_forecast", "path": "/v1/forecast", "method": "GET", "parameters": {}}
		]
	}`)

	manual, err := ParseManual(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manual.Tools[0].GetEndpointURL(manual.BaseURL) != "https://api.weather.com/v1/forecast" {
		t.Errorf("unexpected URL: %s", manual.Tools[0].GetEndpointURL(manual.BaseURL))
	}
}

func TestParseManual_ToolCallTemplateFormat(t *testing.T) {
	data := []byte(`{
		"manual_version": "1.0.0",
		"tools": [{
			"name": "search",
			"description": "Search",
			"inputs": {"type": "object"},
			"tool_call_template": {"call_template_type": "http", "url": "https://api.com/search", "http_method": "GET"}
		}]
	}`)

	manual, err := ParseManual(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url := manual.Tools[0].GetEndpointURL(""); url != "https://api.com/search" {
		t.Errorf("expected tool_call_template URL, got '%s'", url)
	}
}

func TestParseManual_NoTools(t *testing.T) {
	_, err := ParseManual([]byte(`{"name": "test", "tools": []}`))
	if err == nil {
		t.Fatal("expected error for empty tools")
	}
}

func TestParseManual_MalformedJSON(t *testing.T) {
	_, err := ParseManual([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestParseManual_ToolMissingName(t *testing.T) {
	_, err := ParseManual([]byte(`{"tools": [{"description": "no name"}]}`))
	if err == nil {
		t.Fatal("expected error for tool missing name")
	}
}

func TestParseManual_ToolWithoutURL_NoError(t *testing.T) {
	data := []byte(`{"name": "test", "tools": [{"name": "register", "description": "Register"}]}`)
	manual, err := ParseManual(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(manual.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(manual.Tools))
	}
}

func TestConvertToToolRequests_RealFormat(t *testing.T) {
	manual := &UTCPManualJSON{
		Name: "UtcpApi",
		Auth: &UTCPAuthSpec{
			AuthType: "bearer",
			VarName:  "Authorization",
			Location: "header",
			Prefix:   "Bearer",
		},
		Tools: []UTCPToolSpec{
			{
				Name:        "register",
				Description: "Register a user",
				Inputs:      json.RawMessage(`{"type": "object", "properties": {"username": {"type": "string"}}}`),
				ToolProvider: &UTCPToolProvider{
					ProviderType: "http",
					HTTPMethod:   "POST",
					URL:          "http://localhost:5227/auth/register",
				},
			},
			{
				Name:        "list_products",
				Description: "List products",
				Inputs:      json.RawMessage(`{"type": "object"}`),
				ToolProvider: &UTCPToolProvider{
					ProviderType: "http",
					HTTPMethod:   "GET",
					URL:          "http://localhost:5227/products",
					Auth: &UTCPAuthSpec{
						AuthType: "bearer",
						VarName:  "Authorization",
						Location: "header",
					},
				},
			},
		},
	}

	values := map[string]string{"TOKEN": "my-jwt-token"}
	reqs := ConvertToToolRequests(manual, "UtcpApi", values)

	if len(reqs) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(reqs))
	}

	// Check register tool
	if reqs[0].Name != "UtcpApi__register" {
		t.Errorf("expected 'UtcpApi__register', got '%s'", reqs[0].Name)
	}
	if reqs[0].EndpointURL != "http://localhost:5227/auth/register" {
		t.Errorf("expected register URL, got '%s'", reqs[0].EndpointURL)
	}
	if reqs[0].HTTPMethod != "POST" {
		t.Errorf("expected POST, got '%s'", reqs[0].HTTPMethod)
	}
	if reqs[0].AuthType != "bearer" {
		t.Errorf("expected bearer auth, got '%s'", reqs[0].AuthType)
	}

	// Check auth config has token
	var cfg map[string]string
	json.Unmarshal(reqs[0].AuthConfig, &cfg)
	if cfg["token"] != "my-jwt-token" {
		t.Errorf("expected token 'my-jwt-token', got '%s'", cfg["token"])
	}

	// Check list_products
	if reqs[1].EndpointURL != "http://localhost:5227/products" {
		t.Errorf("expected products URL, got '%s'", reqs[1].EndpointURL)
	}
	if reqs[1].HTTPMethod != "GET" {
		t.Errorf("expected GET, got '%s'", reqs[1].HTTPMethod)
	}
}

func TestConvertToToolRequests_NoToken(t *testing.T) {
	manual := &UTCPManualJSON{
		Name: "api",
		Auth: &UTCPAuthSpec{AuthType: "bearer"},
		Tools: []UTCPToolSpec{
			{
				Name: "public_endpoint",
				ToolProvider: &UTCPToolProvider{
					HTTPMethod: "GET",
					URL:        "http://example.com/public",
				},
			},
		},
	}

	reqs := ConvertToToolRequests(manual, "api", nil)
	if len(reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(reqs))
	}
	if reqs[0].EndpointURL != "http://example.com/public" {
		t.Errorf("unexpected URL: %s", reqs[0].EndpointURL)
	}
}

func TestResolveAuth_BearerWithToken(t *testing.T) {
	spec := &UTCPAuthSpec{
		AuthType: "bearer",
		VarName:  "Authorization",
		Location: "header",
		Prefix:   "Bearer",
	}
	values := map[string]string{"TOKEN": "jwt123"}

	authType, authConfig, _ := ResolveAuth(spec, values)
	if authType != "bearer" {
		t.Errorf("expected 'bearer', got '%s'", authType)
	}

	var cfg map[string]string
	json.Unmarshal(authConfig, &cfg)
	if cfg["token"] != "jwt123" {
		t.Errorf("expected 'jwt123', got '%s'", cfg["token"])
	}
}

func TestResolveAuth_BearerNoToken(t *testing.T) {
	spec := &UTCPAuthSpec{AuthType: "bearer"}
	authType, authConfig, _ := ResolveAuth(spec, nil)
	if authType != "bearer" {
		t.Errorf("expected 'bearer', got '%s'", authType)
	}
	var cfg map[string]string
	json.Unmarshal(authConfig, &cfg)
	if cfg["token"] != "" {
		t.Errorf("expected empty token, got '%s'", cfg["token"])
	}
}

func TestResolveAuth_APIKeyHeader(t *testing.T) {
	spec := &UTCPAuthSpec{
		Type:     "api_key",
		Location: "header",
		Vars:     map[string]string{"X-Api-Key": "Your API key"},
	}
	values := map[string]string{"X-Api-Key": "secret123"}

	authType, authConfig, _ := ResolveAuth(spec, values)
	if authType != "api_key" {
		t.Errorf("expected 'api_key', got '%s'", authType)
	}
	var cfg map[string]string
	json.Unmarshal(authConfig, &cfg)
	if cfg["key"] != "secret123" {
		t.Errorf("expected 'secret123', got '%s'", cfg["key"])
	}
}

func TestResolveAuth_APIKeyQuery(t *testing.T) {
	spec := &UTCPAuthSpec{
		AuthType: "api_key",
		APIKey:   "${API_KEY}",
		VarName:  "appid",
		Location: "query",
	}
	values := map[string]string{"API_KEY": "qval"}

	authType, _, queryParams := ResolveAuth(spec, values)
	if authType != "" {
		t.Errorf("expected empty auth_type, got '%s'", authType)
	}
	if queryParams["appid"] != "qval" {
		t.Errorf("expected 'qval', got '%s'", queryParams["appid"])
	}
}

func TestResolveAuth_Nil(t *testing.T) {
	authType, authConfig, qp := ResolveAuth(nil, nil)
	if authType != "" || string(authConfig) != `{}` || qp != nil {
		t.Error("expected empty result for nil spec")
	}
}

func TestGetAuthVars_BearerNoVar(t *testing.T) {
	spec := &UTCPAuthSpec{AuthType: "bearer", VarName: "Authorization"}
	vars := GetAuthVars(spec)
	if len(vars) != 1 || vars[0] != "TOKEN" {
		t.Errorf("expected [TOKEN], got %v", vars)
	}
}

func TestGetAuthVars_APIKeyWithVar(t *testing.T) {
	spec := &UTCPAuthSpec{AuthType: "api_key", APIKey: "${API_KEY}"}
	vars := GetAuthVars(spec)
	if len(vars) != 1 || vars[0] != "API_KEY" {
		t.Errorf("expected [API_KEY], got %v", vars)
	}
}

func TestGetAuthVars_SimplifiedVars(t *testing.T) {
	spec := &UTCPAuthSpec{Type: "api_key", Vars: map[string]string{"X-Key": "d"}}
	vars := GetAuthVars(spec)
	if len(vars) != 1 {
		t.Errorf("expected 1 var, got %d", len(vars))
	}
}

func TestGetAuthVars_Nil(t *testing.T) {
	if GetAuthVars(nil) != nil {
		t.Error("expected nil")
	}
}

func TestSanitizeToolName(t *testing.T) {
	tests := []struct{ in, want string }{
		{"simple", "simple"},
		{"with.dot", "with_dot"},
		{"with spaces", "with_spaces"},
		{"a-b_c", "a-b_c"},
		{"café", "caf_"},
	}
	for _, tt := range tests {
		got := sanitizeToolName(tt.in)
		if got != tt.want {
			t.Errorf("sanitizeToolName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestToolSpec_GetEndpointURL_Priority(t *testing.T) {
	// tool_provider takes priority over tool_call_template over path
	tool := UTCPToolSpec{
		Path:             "/fallback",
		ToolCallTemplate: &UTCPToolCallTemplate{URL: "http://template.com"},
		ToolProvider:     &UTCPToolProvider{URL: "http://provider.com"},
	}
	if url := tool.GetEndpointURL("http://base.com"); url != "http://provider.com" {
		t.Errorf("expected provider URL, got '%s'", url)
	}

	// Without provider, falls to template
	tool2 := UTCPToolSpec{
		Path:             "/fallback",
		ToolCallTemplate: &UTCPToolCallTemplate{URL: "http://template.com"},
	}
	if url := tool2.GetEndpointURL("http://base.com"); url != "http://template.com" {
		t.Errorf("expected template URL, got '%s'", url)
	}

	// Without both, falls to base + path
	tool3 := UTCPToolSpec{Path: "/v1/search"}
	if url := tool3.GetEndpointURL("https://base.com"); url != "https://base.com/v1/search" {
		t.Errorf("expected base+path, got '%s'", url)
	}
}
