package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server     ServerConfig
	DB         DBConfig
	Redis      RedisConfig
	JWT        JWTConfig
	Encryption EncryptionConfig
	XMPP       XMPPConfig
	Log        LogConfig
}

type ServerConfig struct {
	Host string
	Port int
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxConns int32
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type EncryptionConfig struct {
	Key string
}

type XMPPConfig struct {
	Domain string
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() (*Config, error) {
	k := koanf.New(".")

	// Load .env file if it exists (ignore error if missing)
	_ = k.Load(file.Provider(".env"), dotenv.Parser())

	// Load environment variables (override .env)
	err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.ToLower(strings.ReplaceAll(s, "_", "."))
	}), nil)
	if err != nil {
		return nil, fmt.Errorf("loading env vars: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: k.String("server.host"),
			Port: k.Int("server.port"),
		},
		DB: DBConfig{
			Host:     k.String("db.host"),
			Port:     k.Int("db.port"),
			User:     k.String("db.user"),
			Password: k.String("db.password"),
			Name:     k.String("db.name"),
			SSLMode:  k.String("db.sslmode"),
			MaxConns: int32(k.Int("db.max.conns")),
		},
		Redis: RedisConfig{
			Host:     k.String("redis.host"),
			Port:     k.Int("redis.port"),
			Password: k.String("redis.password"),
			DB:       k.Int("redis.db"),
		},
		JWT: JWTConfig{
			AccessSecret:  k.String("jwt.access.secret"),
			RefreshSecret: k.String("jwt.refresh.secret"),
		},
		Encryption: EncryptionConfig{
			Key: k.String("encryption.key"),
		},
		XMPP: XMPPConfig{
			Domain: k.String("xmpp.domain"),
		},
		Log: LogConfig{
			Level:  k.String("log.level"),
			Format: k.String("log.format"),
		},
	}

	// Apply defaults
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.DB.Host == "" {
		cfg.DB.Host = "localhost"
	}
	if cfg.DB.Port == 0 {
		cfg.DB.Port = 5432
	}
	if cfg.DB.User == "" {
		cfg.DB.User = "aiox"
	}
	if cfg.DB.Name == "" {
		cfg.DB.Name = "aiox"
	}
	if cfg.DB.SSLMode == "" {
		cfg.DB.SSLMode = "disable"
	}
	if cfg.DB.MaxConns == 0 {
		cfg.DB.MaxConns = 25
	}
	if cfg.Redis.Host == "" {
		cfg.Redis.Host = "localhost"
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}
	if cfg.XMPP.Domain == "" {
		cfg.XMPP.Domain = "aiox.local"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "debug"
	}
	if cfg.Log.Format == "" {
		cfg.Log.Format = "text"
	}

	// Parse durations
	accessExpStr := k.String("jwt.access.expiry")
	if accessExpStr == "" {
		accessExpStr = "15m"
	}
	cfg.JWT.AccessExpiry, err = time.ParseDuration(accessExpStr)
	if err != nil {
		return nil, fmt.Errorf("parsing jwt access expiry: %w", err)
	}

	refreshExpStr := k.String("jwt.refresh.expiry")
	if refreshExpStr == "" {
		refreshExpStr = "168h"
	}
	cfg.JWT.RefreshExpiry, err = time.ParseDuration(refreshExpStr)
	if err != nil {
		return nil, fmt.Errorf("parsing jwt refresh expiry: %w", err)
	}

	return cfg, nil
}
