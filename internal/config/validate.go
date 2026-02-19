package config

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

// Validate checks Config for production-critical problems.
// It collects all errors into a single joined error.
func (c *Config) Validate() error {
	var errs []string

	// JWT secrets
	if len(c.JWT.AccessSecret) < 32 {
		errs = append(errs, "JWT_ACCESS_SECRET must be at least 32 characters")
	}
	if len(c.JWT.RefreshSecret) < 32 {
		errs = append(errs, "JWT_REFRESH_SECRET must be at least 32 characters")
	}
	if c.JWT.AccessSecret != "" && c.JWT.RefreshSecret != "" && c.JWT.AccessSecret == c.JWT.RefreshSecret {
		errs = append(errs, "JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must differ")
	}

	// Encryption key: must be exactly 64 hex chars (32 bytes)
	if c.Encryption.Key == "" {
		errs = append(errs, "ENCRYPTION_KEY is required")
	} else if len(c.Encryption.Key) != 64 {
		errs = append(errs, "ENCRYPTION_KEY must be exactly 64 hex characters (32 bytes)")
	} else if _, err := hex.DecodeString(c.Encryption.Key); err != nil {
		errs = append(errs, "ENCRYPTION_KEY must be valid hex")
	}

	// DB password
	if c.DB.Password == "" {
		errs = append(errs, "DB_PASSWORD is required")
	}

	// Port ranges
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		errs = append(errs, fmt.Sprintf("SERVER_PORT must be 1–65535, got %d", c.Server.Port))
	}
	if c.DB.Port < 1 || c.DB.Port > 65535 {
		errs = append(errs, fmt.Sprintf("DB_PORT must be 1–65535, got %d", c.DB.Port))
	}
	if c.Redis.Port < 1 || c.Redis.Port > 65535 {
		errs = append(errs, fmt.Sprintf("REDIS_PORT must be 1–65535, got %d", c.Redis.Port))
	}
	if c.GRPC.Port < 1 || c.GRPC.Port > 65535 {
		errs = append(errs, fmt.Sprintf("GRPC_PORT must be 1–65535, got %d", c.GRPC.Port))
	}

	// Worker API key: warn only
	if c.GRPC.WorkerAPIKey == "" {
		slog.Warn("GRPC_WORKER_API_KEY is empty — gRPC server has no authentication")
	}

	if len(errs) > 0 {
		return errors.New("config validation failed:\n  " + strings.Join(errs, "\n  "))
	}
	return nil
}
