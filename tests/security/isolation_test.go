//go:build integration

package security

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/aiox-platform/aiox/internal/agents"
	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
	"github.com/aiox-platform/aiox/internal/users"
)

type testEnv struct {
	server *httptest.Server
	pool   *pgxpool.Pool
}

func setupSecurityTestEnv(t *testing.T) *testEnv {
	t.Helper()
	ctx := context.Background()

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "pgvector/pgvector:0.8.1-pg16",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
				"POSTGRES_DB":       "aiox_security_test",
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { pgContainer.Terminate(ctx) })

	pgHost, _ := pgContainer.Host(ctx)
	pgPort, _ := pgContainer.MappedPort(ctx, "5432")

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { redisContainer.Terminate(ctx) })

	redisHost, _ := redisContainer.Host(ctx)
	redisPort, _ := redisContainer.MappedPort(ctx, "6379")

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/aiox_security_test?sslmode=disable", pgHost, pgPort.Port())
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { pool.Close() })

	_, err = pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; CREATE EXTENSION IF NOT EXISTS "vector";`)
	require.NoError(t, err)

	migrationsPath := getMigrationsPath()
	m, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), dsn)
	require.NoError(t, err)
	require.NoError(t, m.Up())

	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort.Port()),
	})
	t.Cleanup(func() { redisClient.Close() })

	encKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	jwtMgr := auth.NewJWTManager("sec-test-access-secret-32-chars!!", "sec-test-refresh-secret-32-chars!!", 15*time.Minute, 7*24*time.Hour)
	authSvc := auth.NewService(jwtMgr, redisClient)
	userRepo := users.NewRepository(pool)
	userSvc := users.NewService(userRepo)
	authHandler := auth.NewHandler(authSvc, userSvc)

	agentRepo := agents.NewRepository(pool)
	agentSvc := agents.NewService(agentRepo, encKey, "security.test")
	agentHandler := agents.NewHandler(agentSvc)

	router := api.NewRouter(pool, api.HandlerSet{
		Register:            authHandler.Register,
		Login:               authHandler.Login,
		Refresh:             authHandler.Refresh,
		Logout:              authHandler.Logout,
		CreateAgent:         agentHandler.Create,
		ListAgents:          agentHandler.List,
		GetAgent:            agentHandler.Get,
		UpdateAgent:         agentHandler.Update,
		DeleteAgent:         agentHandler.Delete,
		OwnershipMiddleware: agentHandler.OwnershipMiddleware,
		AuthMiddleware:      auth.Middleware(authSvc),
	})

	server := httptest.NewServer(router)
	t.Cleanup(func() { server.Close() })

	return &testEnv{server: server, pool: pool}
}

func getMigrationsPath() string {
	paths := []string{"../../migrations", "../../../migrations"}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	log.Fatal("migrations directory not found")
	return ""
}

func doReq(t *testing.T, env *testEnv, method, path string, body any, token string) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, env.server.URL+path, r)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func parseResp(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var m map[string]any
	json.NewDecoder(resp.Body).Decode(&m)
	return m
}

func register(t *testing.T, env *testEnv, email string) string {
	t.Helper()
	resp := doReq(t, env, "POST", "/api/v1/auth/register", map[string]string{"email": email, "password": "password123"}, "")
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	r := parseResp(t, resp)
	return r["data"].(map[string]any)["access_token"].(string)
}

// TestMultiTenantBoundary tests that multi-tenant isolation is enforced across
// many users trying to access each other's resources.
func TestMultiTenantBoundary(t *testing.T) {
	env := setupSecurityTestEnv(t)

	// Create 5 users, each with an agent
	type userAgent struct {
		token   string
		agentID string
	}

	var uas []userAgent
	for i := 0; i < 5; i++ {
		email := fmt.Sprintf("tenant-%d@security.test", i)
		token := register(t, env, email)

		body := map[string]any{
			"name":          fmt.Sprintf("Agent %d", i),
			"system_prompt": fmt.Sprintf("Secret prompt for tenant %d", i),
		}
		resp := doReq(t, env, "POST", "/api/v1/agents", body, token)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		result := parseResp(t, resp)
		agentID := result["data"].(map[string]any)["id"].(string)
		uas = append(uas, userAgent{token: token, agentID: agentID})
	}

	t.Run("no user can access another users agent", func(t *testing.T) {
		for i, ua := range uas {
			for j, other := range uas {
				if i == j {
					continue
				}
				// Try GET
				resp := doReq(t, env, "GET", "/api/v1/agents/"+other.agentID, nil, ua.token)
				assert.Equal(t, http.StatusForbidden, resp.StatusCode,
					"user %d should not GET user %d's agent", i, j)
				resp.Body.Close()

				// Try PUT
				resp = doReq(t, env, "PUT", "/api/v1/agents/"+other.agentID,
					map[string]any{"name": "hacked"}, ua.token)
				assert.Equal(t, http.StatusForbidden, resp.StatusCode,
					"user %d should not UPDATE user %d's agent", i, j)
				resp.Body.Close()

				// Try DELETE
				resp = doReq(t, env, "DELETE", "/api/v1/agents/"+other.agentID, nil, ua.token)
				assert.Equal(t, http.StatusForbidden, resp.StatusCode,
					"user %d should not DELETE user %d's agent", i, j)
				resp.Body.Close()
			}
		}
	})

	t.Run("each user only sees own agents in list", func(t *testing.T) {
		for i, ua := range uas {
			resp := doReq(t, env, "GET", "/api/v1/agents", nil, ua.token)
			require.Equal(t, http.StatusOK, resp.StatusCode)

			result := parseResp(t, resp)
			agentsList := result["data"].([]any)

			for _, a := range agentsList {
				agent := a.(map[string]any)
				assert.Equal(t, ua.agentID, agent["id"].(string),
					"user %d should only see their own agent", i)
			}
		}
	})

	t.Run("system_prompt encrypted at rest", func(t *testing.T) {
		for i, ua := range uas {
			var profileBytes []byte
			err := env.pool.QueryRow(context.Background(),
				"SELECT profile FROM agents WHERE id = $1", ua.agentID).Scan(&profileBytes)
			require.NoError(t, err)

			var profile map[string]any
			require.NoError(t, json.Unmarshal(profileBytes, &profile))

			plaintext := fmt.Sprintf("Secret prompt for tenant %d", i)
			assert.NotEqual(t, plaintext, profile["system_prompt"],
				"tenant %d system_prompt should be encrypted in DB", i)
		}
	})
}
