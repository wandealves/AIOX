package config

import (
	"strings"
	"testing"
	"time"
)

func validConfig() *Config {
	return &Config{
		Server: ServerConfig{Host: "0.0.0.0", Port: 8080},
		DB: DBConfig{
			Host: "localhost", Port: 5432, User: "aiox",
			Password: "secret", Name: "aiox", SSLMode: "disable", MaxConns: 25,
		},
		Redis: RedisConfig{Host: "localhost", Port: 6379},
		JWT: JWTConfig{
			AccessSecret:  "access-secret-that-is-at-least-32-chars!",
			RefreshSecret: "refresh-secret-that-is-at-least-32-chr!",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 168 * time.Hour,
		},
		Encryption: EncryptionConfig{Key: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"},
		GRPC:       GRPCConfig{Host: "0.0.0.0", Port: 50051, WorkerAPIKey: "some-key"},
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := validConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestValidate_JWTAccessSecretTooShort(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.AccessSecret = "short"
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "JWT_ACCESS_SECRET") {
		t.Fatalf("expected JWT_ACCESS_SECRET error, got: %v", err)
	}
}

func TestValidate_JWTRefreshSecretTooShort(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.RefreshSecret = "short"
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "JWT_REFRESH_SECRET") {
		t.Fatalf("expected JWT_REFRESH_SECRET error, got: %v", err)
	}
}

func TestValidate_JWTSecretsMustDiffer(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.AccessSecret = "the-same-secret-that-is-at-least-32-chars!"
	cfg.JWT.RefreshSecret = "the-same-secret-that-is-at-least-32-chars!"
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("expected 'must differ' error, got: %v", err)
	}
}

func TestValidate_EncryptionKeyRequired(t *testing.T) {
	cfg := validConfig()
	cfg.Encryption.Key = ""
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "ENCRYPTION_KEY is required") {
		t.Fatalf("expected ENCRYPTION_KEY required error, got: %v", err)
	}
}

func TestValidate_EncryptionKeyWrongLength(t *testing.T) {
	cfg := validConfig()
	cfg.Encryption.Key = "tooshort"
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "64 hex characters") {
		t.Fatalf("expected 64 hex characters error, got: %v", err)
	}
}

func TestValidate_EncryptionKeyInvalidHex(t *testing.T) {
	cfg := validConfig()
	cfg.Encryption.Key = "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "valid hex") {
		t.Fatalf("expected valid hex error, got: %v", err)
	}
}

func TestValidate_DBPasswordRequired(t *testing.T) {
	cfg := validConfig()
	cfg.DB.Password = ""
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "DB_PASSWORD") {
		t.Fatalf("expected DB_PASSWORD error, got: %v", err)
	}
}

func TestValidate_InvalidPorts(t *testing.T) {
	cfg := validConfig()
	cfg.Server.Port = 0
	cfg.DB.Port = 99999
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected port validation errors")
	}
	if !strings.Contains(err.Error(), "SERVER_PORT") {
		t.Errorf("expected SERVER_PORT error in: %v", err)
	}
	if !strings.Contains(err.Error(), "DB_PORT") {
		t.Errorf("expected DB_PORT error in: %v", err)
	}
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Port: 0},
		DB:     DBConfig{Port: 5432},
		Redis:  RedisConfig{Port: 6379},
		GRPC:   GRPCConfig{Port: 50051},
	}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected multiple validation errors")
	}
	// Should contain at least JWT + encryption + DB password + server port errors
	errStr := err.Error()
	for _, substr := range []string{"JWT_ACCESS_SECRET", "JWT_REFRESH_SECRET", "ENCRYPTION_KEY", "DB_PASSWORD", "SERVER_PORT"} {
		if !strings.Contains(errStr, substr) {
			t.Errorf("expected %q in error: %s", substr, errStr)
		}
	}
}
