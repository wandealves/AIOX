//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/aiox-platform/aiox/internal/agents"
	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
	"github.com/aiox-platform/aiox/internal/memory"
	"github.com/aiox-platform/aiox/internal/users"
)

type TestEnv struct {
	Pool        *pgxpool.Pool
	RedisClient *redis.Client
	Server      *httptest.Server
	AuthSvc     *auth.Service
	UserSvc     *users.Service
}

var testEnv *TestEnv

func SetupTestEnv(t *testing.T) *TestEnv {
	t.Helper()
	if testEnv != nil {
		return testEnv
	}

	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "pgvector/pgvector:0.8.1-pg16",
			ExposedPorts: []string{"5433/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "aiox_test",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("starting postgres container: %v", err)
	}
	t.Cleanup(func() { pgContainer.Terminate(ctx) })

	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "3")

	// Start Redis container
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("starting redis container: %v", err)
	}
	t.Cleanup(func() { redisContainer.Terminate(ctx) })

	redisHost, _ := redisContainer.Host(ctx)
	redisPort, _ := redisContainer.MappedPort(ctx, "6379")

	// Connect to PostgreSQL
	dsn := fmt.Sprintf("postgres://test:test@%s:%s/aiox_test?sslmode=disable", pgHost, pgPort.Port())
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connecting to postgres: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	// Enable extensions
	_, err = pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; CREATE EXTENSION IF NOT EXISTS "vector";`)
	if err != nil {
		t.Fatalf("enabling extensions: %v", err)
	}

	// Run migrations
	migrationsPath := getMigrationsPath()
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dsn,
	)
	if err != nil {
		t.Fatalf("creating migrator: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("running migrations: %v", err)
	}

	// Connect to Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort.Port()),
	})
	t.Cleanup(func() { redisClient.Close() })

	// Setup services
	encryptionKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	xmppDomain := "test.aiox.local"

	jwtManager := auth.NewJWTManager("test-access-secret-32-chars-long!!", "test-refresh-secret-32-chars-long!!", 15*time.Minute, 7*24*time.Hour)
	authSvc := auth.NewService(jwtManager, redisClient)
	userRepo := users.NewRepository(pool)
	userSvc := users.NewService(userRepo)
	authHandler := auth.NewHandler(authSvc, userSvc)

	agentRepo := agents.NewRepository(pool)
	agentSvc := agents.NewService(agentRepo, encryptionKey, xmppDomain)
	agentHandler := agents.NewHandler(agentSvc)

	// Memory (Phase 4)
	memoryRepo := memory.NewPostgresRepository(pool)
	shortTermStore := memory.NewShortTermStore(redisClient)
	memorySvc := memory.NewService(memoryRepo, shortTermStore)
	memoryHandler := memory.NewHandler(memorySvc)

	router := api.NewRouter(pool, nil, api.HandlerSet{
		Register: authHandler.Register,
		Login:    authHandler.Login,
		Refresh:  authHandler.Refresh,
		Logout:   authHandler.Logout,

		CreateAgent:         agentHandler.Create,
		ListAgents:          agentHandler.List,
		GetAgent:            agentHandler.Get,
		UpdateAgent:         agentHandler.Update,
		DeleteAgent:         agentHandler.Delete,
		OwnershipMiddleware: agentHandler.OwnershipMiddleware,

		ListMemories:      memoryHandler.List,
		CreateMemory:      memoryHandler.Create,
		SearchMemories:    memoryHandler.Search,
		DeleteMemory:      memoryHandler.Delete,
		DeleteAllMemories: memoryHandler.DeleteAll,

		AuthMiddleware: auth.Middleware(authSvc),
	})

	server := httptest.NewServer(router)
	t.Cleanup(func() { server.Close() })

	testEnv = &TestEnv{
		Pool:        pool,
		RedisClient: redisClient,
		Server:      server,
		AuthSvc:     authSvc,
		UserSvc:     userSvc,
	}

	return testEnv
}

func getMigrationsPath() string {
	// Try relative paths from test directory
	paths := []string{
		"../../migrations",
		"../../../migrations",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	log.Fatal("migrations directory not found")
	return ""
}

// Helper functions

func RegisterUser(t *testing.T, env *TestEnv, email, password string) map[string]any {
	t.Helper()
	body := map[string]string{"email": email, "password": password}
	resp := DoRequest(t, env, "POST", "/api/v1/auth/register", body, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed: status %d", resp.StatusCode)
	}
	return ParseResponse(t, resp)
}

func LoginUser(t *testing.T, env *TestEnv, email, password string) string {
	t.Helper()
	body := map[string]string{"email": email, "password": password}
	resp := DoRequest(t, env, "POST", "/api/v1/auth/login", body, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: status %d", resp.StatusCode)
	}
	result := ParseResponse(t, resp)
	data := result["data"].(map[string]any)
	return data["access_token"].(string)
}

func DoRequest(t *testing.T, env *TestEnv, method, path string, body any, token string) *http.Response {
	t.Helper()
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, env.Server.URL+path, bodyReader)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("doing request: %v", err)
	}
	return resp
}

func ParseResponse(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("parsing response: %v", err)
	}
	return result
}
