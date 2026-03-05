package utcp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/aiox-platform/aiox/internal/tools"
	"github.com/google/uuid"
)

// varPattern matches ${VAR_NAME} placeholders in UTCP auth fields.
var varPattern = regexp.MustCompile(`\$\{(\w+)\}`)

// invalidToolNameChars matches any character not allowed in LLM tool names.
var invalidToolNameChars = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// sanitizeToolName replaces invalid characters with underscores.
// LLM providers (OpenAI, Anthropic) only allow ^[a-zA-Z0-9_-]+$.
func sanitizeToolName(name string) string {
	return invalidToolNameChars.ReplaceAllString(name, "_")
}

// ParseManual validates and unmarshals raw UTCP manual JSON.
func ParseManual(data []byte) (*UTCPManualJSON, error) {
	var manual UTCPManualJSON
	if err := json.Unmarshal(data, &manual); err != nil {
		return nil, fmt.Errorf("invalid UTCP manual JSON: %w", err)
	}

	if len(manual.Tools) == 0 {
		return nil, fmt.Errorf("UTCP manual has no tools defined")
	}

	if manual.Version == "" && manual.ManualVersion != "" {
		manual.Version = manual.ManualVersion
	}

	for i, t := range manual.Tools {
		if t.Name == "" {
			return nil, fmt.Errorf("tool at index %d missing name", i)
		}
		// Default method for simplified format
		if t.Method == "" && t.ToolProvider == nil && t.ToolCallTemplate == nil {
			manual.Tools[i].Method = "POST"
		}
	}

	return &manual, nil
}

// ResolveAuth maps UTCP auth spec + user-provided values to AIOX auth_type and auth_config.
func ResolveAuth(spec *UTCPAuthSpec, values map[string]string) (authType string, authConfig json.RawMessage, queryParams map[string]string) {
	if spec == nil {
		return "", json.RawMessage(`{}`), nil
	}

	effectiveType := spec.GetAuthType()

	switch effectiveType {
	case "api_key":
		return resolveAPIKeyAuth(spec, values)
	case "basic":
		return resolveBasicAuth(spec, values)
	case "bearer":
		return resolveBearerAuth(spec, values)
	}

	return "", json.RawMessage(`{}`), nil
}

func resolveAPIKeyAuth(spec *UTCPAuthSpec, values map[string]string) (string, json.RawMessage, map[string]string) {
	// Official format: api_key="${VAR_NAME}", var_name="param_name", location="query|header"
	if spec.APIKey != "" {
		resolvedKey := resolveVariable(spec.APIKey, values)
		paramName := spec.VarName
		if paramName == "" {
			paramName = "api_key"
		}

		if spec.Location == "query" {
			return "", json.RawMessage(`{}`), map[string]string{paramName: resolvedKey}
		}
		cfg, _ := json.Marshal(map[string]string{
			"key":    resolvedKey,
			"header": paramName,
		})
		return "api_key", cfg, nil
	}

	// Simplified format: vars map
	resolved := resolveVars(spec.Vars, values)

	if spec.Location == "query" {
		qp := make(map[string]string)
		for varName := range spec.Vars {
			if val, ok := resolved[varName]; ok {
				qp[varName] = val
			}
		}
		return "", json.RawMessage(`{}`), qp
	}

	var headerName, keyValue string
	for varName, val := range resolved {
		headerName = varName
		keyValue = val
		break
	}
	cfg, _ := json.Marshal(map[string]string{
		"key":    keyValue,
		"header": headerName,
	})
	return "api_key", cfg, nil
}

func resolveBasicAuth(spec *UTCPAuthSpec, values map[string]string) (string, json.RawMessage, map[string]string) {
	resolved := resolveVars(spec.Vars, values)

	var user, pass string
	for varName, val := range resolved {
		lower := strings.ToLower(varName)
		if strings.Contains(lower, "user") || strings.Contains(lower, "username") {
			user = val
		} else if strings.Contains(lower, "pass") || strings.Contains(lower, "password") {
			pass = val
		}
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	cfg, _ := json.Marshal(map[string]string{"token": encoded})
	return "bearer", cfg, nil
}

func resolveBearerAuth(spec *UTCPAuthSpec, values map[string]string) (string, json.RawMessage, map[string]string) {
	// Official format with ${VAR}
	if spec.APIKey != "" {
		token := resolveVariable(spec.APIKey, values)
		cfg, _ := json.Marshal(map[string]string{"token": token})
		return "bearer", cfg, nil
	}

	// Direct bearer: user provides token via auth_values["TOKEN"]
	if token, ok := values["TOKEN"]; ok && token != "" {
		cfg, _ := json.Marshal(map[string]string{"token": token})
		return "bearer", cfg, nil
	}

	// Simplified format with vars
	resolved := resolveVars(spec.Vars, values)
	for _, val := range resolved {
		if val != "" {
			cfg, _ := json.Marshal(map[string]string{"token": val})
			return "bearer", cfg, nil
		}
	}

	// No token provided — return bearer with empty token (tools may not need auth, e.g. register/login)
	cfg, _ := json.Marshal(map[string]string{"token": ""})
	return "bearer", cfg, nil
}

// ConvertToToolRequests converts parsed UTCP tools into AIOX CreateToolRequest objects.
func ConvertToToolRequests(manual *UTCPManualJSON, manualName string, values map[string]string) []tools.CreateToolRequest {
	// Resolve global auth
	globalAuthType, globalAuthConfig, globalQueryParams := ResolveAuth(manual.Auth, values)

	var result []tools.CreateToolRequest
	for _, t := range manual.Tools {
		namespacedName := sanitizeToolName(manualName) + "__" + sanitizeToolName(t.Name)

		endpointURL := t.GetEndpointURL(manual.BaseURL)

		// Determine auth for this tool
		authType := globalAuthType
		authConfig := globalAuthConfig
		queryParams := globalQueryParams

		// Per-tool auth override
		if toolAuth := t.GetToolAuth(); toolAuth != nil {
			authType, authConfig, queryParams = ResolveAuth(toolAuth, values)
			// If per-tool auth is bearer but has no token, fall back to global
			if toolAuth.GetAuthType() == "bearer" {
				var cfg map[string]string
				_ = json.Unmarshal(authConfig, &cfg)
				if cfg["token"] == "" {
					authType, authConfig, queryParams = ResolveAuth(manual.Auth, values)
				}
			}
		}

		if len(queryParams) > 0 {
			endpointURL = appendQueryParams(endpointURL, queryParams)
		}

		method := strings.ToUpper(t.GetHTTPMethod())

		params := t.GetParameters()

		headers := t.Headers
		if len(headers) == 0 {
			headers = json.RawMessage(`{}`)
		}

		timeoutSec := t.TimeoutSec
		if timeoutSec <= 0 {
			timeoutSec = 30
		}

		result = append(result, tools.CreateToolRequest{
			Name:        namespacedName,
			Description: t.Description,
			Parameters:  params,
			EndpointURL: endpointURL,
			HTTPMethod:  method,
			Headers:     headers,
			AuthType:    authType,
			AuthConfig:  authConfig,
			TimeoutSec:  timeoutSec,
		})
	}
	return result
}

// GetAuthVars extracts variable names that the user needs to provide.
func GetAuthVars(spec *UTCPAuthSpec) []string {
	if spec == nil {
		return nil
	}

	// Official format: extract ${VAR_NAME} from api_key field
	if spec.APIKey != "" {
		matches := varPattern.FindAllStringSubmatch(spec.APIKey, -1)
		vars := make([]string, 0, len(matches))
		for _, m := range matches {
			if len(m) > 1 {
				vars = append(vars, m[1])
			}
		}
		if len(vars) > 0 {
			return vars
		}
	}

	// Simplified format: vars map keys
	if len(spec.Vars) > 0 {
		vars := make([]string, 0, len(spec.Vars))
		for k := range spec.Vars {
			vars = append(vars, k)
		}
		return vars
	}

	// Bearer without ${VAR}: user needs to provide a TOKEN
	if spec.GetAuthType() == "bearer" {
		return []string{"TOKEN"}
	}

	return nil
}

func resolveVariable(template string, values map[string]string) string {
	return varPattern.ReplaceAllStringFunc(template, func(match string) string {
		varName := match[2 : len(match)-1]
		if val, ok := values[varName]; ok {
			return val
		}
		return match
	})
}

func resolveVars(vars map[string]string, values map[string]string) map[string]string {
	resolved := make(map[string]string, len(vars))
	for varName := range vars {
		if val, ok := values[varName]; ok {
			resolved[varName] = val
		}
	}
	return resolved
}

func appendQueryParams(url string, queryParams map[string]string) string {
	if len(queryParams) == 0 {
		return url
	}
	parts := make([]string, 0, len(queryParams))
	for k, v := range queryParams {
		parts = append(parts, k+"="+v)
	}
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	return url + sep + strings.Join(parts, "&")
}

// CreateToolFromRequest creates a ToolDefinition from a CreateToolRequest with UTCP-specific fields.
func CreateToolFromRequest(agentID uuid.UUID, manualID uuid.UUID, req tools.CreateToolRequest) *tools.ToolDefinition {
	return &tools.ToolDefinition{
		ID:           uuid.New(),
		AgentID:      agentID,
		Name:         req.Name,
		Description:  req.Description,
		Parameters:   req.Parameters,
		EndpointURL:  req.EndpointURL,
		HTTPMethod:   req.HTTPMethod,
		Headers:      req.Headers,
		AuthType:     req.AuthType,
		AuthConfig:   req.AuthConfig,
		IsActive:     true,
		TimeoutSec:   req.TimeoutSec,
		Version:      "1.0.0",
		Category:     "utcp",
		OutputSchema: json.RawMessage(`{}`),
		ToolType:     "utcp",
		UTCPManualID: &manualID,
	}
}
