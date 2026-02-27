package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ValidateToolInput validates a tool's input against its JSON Schema parameters.
// This is a lightweight validation that checks required fields and basic types
// without requiring a full JSON Schema library dependency.
func ValidateToolInput(parameters json.RawMessage, input json.RawMessage) error {
	if len(parameters) == 0 || string(parameters) == "{}" {
		return nil // no schema defined, skip validation
	}
	if len(input) == 0 {
		input = json.RawMessage(`{}`)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(parameters, &schema); err != nil {
		return fmt.Errorf("invalid parameter schema: %w", err)
	}

	var inputMap map[string]interface{}
	if err := json.Unmarshal(input, &inputMap); err != nil {
		return fmt.Errorf("invalid input JSON: %w", err)
	}

	// Check required fields
	if required, ok := schema["required"]; ok {
		if reqArr, ok := required.([]interface{}); ok {
			for _, r := range reqArr {
				if fieldName, ok := r.(string); ok {
					if _, exists := inputMap[fieldName]; !exists {
						return fmt.Errorf("missing required field: %s", fieldName)
					}
				}
			}
		}
	}

	// Check property types if properties are defined
	if props, ok := schema["properties"]; ok {
		if propsMap, ok := props.(map[string]interface{}); ok {
			for key, val := range inputMap {
				propDef, exists := propsMap[key]
				if !exists {
					continue // extra fields are allowed
				}
				if propDefMap, ok := propDef.(map[string]interface{}); ok {
					if err := validateType(key, val, propDefMap); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

// ValidateToolOutput validates a tool's output against its output schema.
func ValidateToolOutput(outputSchema json.RawMessage, output json.RawMessage) error {
	if len(outputSchema) == 0 || string(outputSchema) == "{}" {
		return nil
	}
	if len(output) == 0 {
		return fmt.Errorf("output is empty but schema is defined")
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(outputSchema, &schema); err != nil {
		return fmt.Errorf("invalid output schema: %w", err)
	}

	var outputData interface{}
	if err := json.Unmarshal(output, &outputData); err != nil {
		return fmt.Errorf("invalid output JSON: %w", err)
	}

	schemaType, _ := schema["type"].(string)
	if schemaType != "" {
		if err := checkType(schemaType, outputData); err != nil {
			return fmt.Errorf("output type mismatch: %w", err)
		}
	}

	return nil
}

func validateType(fieldName string, value interface{}, propDef map[string]interface{}) error {
	expectedType, ok := propDef["type"].(string)
	if !ok {
		return nil
	}
	return checkTypeForField(fieldName, expectedType, value)
}

func checkTypeForField(fieldName, expectedType string, value interface{}) error {
	err := checkType(expectedType, value)
	if err != nil {
		return fmt.Errorf("field '%s': %w", fieldName, err)
	}
	return nil
}

func checkType(expectedType string, value interface{}) error {
	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case "number", "integer":
		switch value.(type) {
		case float64, int, int64:
			// ok
		default:
			return fmt.Errorf("expected %s, got %T", expectedType, value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case "array":
		if _, ok := value.([]interface{}); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case "object":
		if _, ok := value.(map[string]interface{}); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	}
	return nil
}

// NormalizeToolResponse wraps a raw tool response in the standard format.
func NormalizeToolResponse(result string, durationMs int, err string) string {
	status := "success"
	var data interface{}
	var errMsg interface{}

	if err != "" {
		status = "error"
		errMsg = err
	} else {
		errMsg = nil
	}

	// Try to parse result as JSON
	if result != "" {
		var parsed interface{}
		if jsonErr := json.Unmarshal([]byte(result), &parsed); jsonErr == nil {
			data = parsed
		} else {
			data = result
		}
	}

	response := map[string]interface{}{
		"status":            status,
		"data":              data,
		"error":             errMsg,
		"execution_time_ms": durationMs,
	}

	b, _ := json.Marshal(response)
	return string(b)
}

// ValidateToolName checks that a tool name is valid.
func ValidateToolName(name string) error {
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if len(name) > 100 {
		return fmt.Errorf("tool name cannot exceed 100 characters")
	}
	// Only alphanumeric, hyphens, underscores
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return fmt.Errorf("tool name contains invalid character: %c", c)
		}
	}
	return nil
}

// ValidateVersion checks that a version string follows semver format.
func ValidateVersion(version string) error {
	if version == "" {
		return nil
	}
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return fmt.Errorf("version must be in semver format (e.g., 1.0.0)")
	}
	return nil
}
