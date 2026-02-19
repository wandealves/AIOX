# AIOX — Multi-Agent AI Platform

AIOX is a production-ready multi-agent AI platform built in Go. Users interact with AI agents over **XMPP** (or REST); agents run in isolated **Python workers** connected via **gRPC**, with **NATS JetStream** as the async message bus, **PostgreSQL** for persistence and vector memory, and **Redis** for caching and rate limiting.

---

## Table of Contents

- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Quick Start (Docker)](#quick-start-docker)
- [TLS Setup for XMPP Clients](#tls-setup-for-xmpp-clients)
- [Registering XMPP Users](#registering-xmpp-users)
- [Local Development](#local-development)
- [Configuration Reference](#configuration-reference)
- [REST API Reference](#rest-api-reference)
- [Using XMPP to Chat with Agents](#using-xmpp-to-chat-with-agents)
- [Python Worker](#python-worker)
- [Make Targets](#make-targets)
- [Running Tests](#running-tests)
- [Project Structure](#project-structure)
- [Troubleshooting](#troubleshooting)

---

## Architecture

```
User ──XMPP──► ejabberd ──► XMPP Component ──► NATS (inbound)
                                                      │
                                               Orchestrator
                                          (validate · route · quota)
                                                      │
                                             NATS (aiox.tasks.*)
                                                      │
                                               Dispatcher
                                       (agent fetch · memory context)
                                                      │
                                         gRPC ──► Python Worker
                                          (OpenAI · Anthropic · Ollama)
                                                      │
                                             NATS (outbound)
                                                      │
                                       Outbound Relay ──► XMPP ──► User
```

### Stack

| Layer | Technology |
|---|---|
| HTTP API | Go + chi |
| Async messaging | NATS JetStream |
| XMPP server | ejabberd |
| AI workers | Python 3.12 + gRPC |
| Database | PostgreSQL 16 + pgvector |
| Cache / Rate limiting | Redis 7 |
| Metrics | Prometheus |

---

## Prerequisites

| Tool | Minimum version | Purpose |
|---|---|---|
| Docker | 24+ | All services |
| Docker Compose | v2 | Orchestration |
| Go | 1.24 | Local dev / unit tests |
| Python | 3.12 | Worker local dev |
| `openssl` | any | TLS certificate generation |
| An XMPP client | — | Chat with agents (e.g. [Dino](https://dino.im)) |

> **Go path note:** if Go is installed at `~/go-sdk/go/bin`, add it to your PATH:
> ```bash
> export PATH=$HOME/go-sdk/go/bin:$PATH
> ```

---

## Quick Start (Docker)

### 1. Clone and configure

```bash
git clone https://github.com/aiox-platform/aiox.git
cd aiox
cp .env.example .env
```

Edit `.env` and set at minimum:

```bash
# Must be ≥32 chars, the two values must differ
JWT_ACCESS_SECRET=your-access-secret-at-least-32-characters
JWT_REFRESH_SECRET=your-refresh-secret-at-least-32-characters

# 64 hex chars = 32-byte AES-256 key
ENCRYPTION_KEY=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef

# Must be ≥32 chars
GRPC_WORKER_API_KEY=your-worker-api-key-at-least-32-characters

# At least one LLM provider
OPENAI_API_KEY=sk-...
# or
ANTHROPIC_API_KEY=sk-ant-...
```

### 2. Add aiox.local to /etc/hosts

```bash
echo "127.0.0.1 aiox.local" | sudo tee -a /etc/hosts
echo "127.0.0.1 agents.aiox.local" | sudo tee -a /etc/hosts
```

### 3. Generate TLS certificates (required for XMPP clients)

```bash
bash docker/ejabberd/gen-cert.sh
sudo bash docker/ejabberd/install-ca.sh   # installs CA into Ubuntu trust store
```

### 4. Start all services

```bash
make up
# or: docker compose up -d
```

Services started:

| Service | Port(s) | Description |
|---|---|---|
| PostgreSQL | 5433 | Primary database |
| Redis | 6379 | Cache + rate limiting |
| NATS | 4222, 8222 | Message bus (HTTP monitor at :8222) |
| ejabberd | 5222, 5275, 5280 | XMPP server |
| aiox-api | 8080, 50051 | REST API + gRPC |
| aiox-worker | — | Python AI worker |

### 5. Verify all services are healthy

```bash
docker compose ps
curl http://localhost:8080/health/ready
```

Expected response:

```json
{
  "status": "ok",
  "database": "ok",
  "nats": "ok",
  "workers": 1
}
```

---

## TLS Setup for XMPP Clients

Modern XMPP clients (Dino, Gajim, Conversations) require a trusted TLS certificate. The `gen-cert.sh` script creates a local CA and a server certificate signed by it.

```bash
# Step 1 — generate (no sudo needed)
bash docker/ejabberd/gen-cert.sh

# Step 2 — install CA into system trust store (Ubuntu/Debian)
sudo bash docker/ejabberd/install-ca.sh

# Step 3 — restart ejabberd to load the new certificate
docker compose restart ejabberd

# Step 4 — confirm ejabberd loaded it (look for "Listening for s2s" or similar)
docker compose logs ejabberd --tail=30
```

> **Note:** The generated files are stored in `docker/ejabberd/certs/` and are excluded from git via `.gitignore`.

---

## Registering XMPP Users

Users must be registered on the ejabberd server to chat with agents over XMPP.

```bash
# Register a user
docker exec -it aiox-ejabberd ejabberdctl register <username> aiox.local <password>

# Example
docker exec -it aiox-ejabberd ejabberdctl register wande aiox.local senha123

# List registered users
docker exec -it aiox-ejabberd ejabberdctl registered_users aiox.local

# Delete a user
docker exec -it aiox-ejabberd ejabberdctl unregister <username> aiox.local
```

### Connecting with Dino

1. Open Dino → **Add Account**
2. JID: `wande@aiox.local`
3. Password: `senha123`
4. Dino will auto-discover the server at `aiox.local:5222`
5. The certificate should now be trusted (after `install-ca.sh`)

---

## Local Development

### 1. Start infrastructure only

```bash
docker compose up -d postgres redis nats ejabberd
```

### 2. Copy and edit config

```bash
cp .env.example .env
# Edit .env with your secrets (see Configuration Reference)
```

### 3. Run database migrations

```bash
make migrate-up
# or: DB_AUTO_MIGRATE=true go run ./cmd/api
```

### 4. Run the API server

```bash
make dev
# or: go run ./cmd/api
```

### 5. Run the Python worker

```bash
cd worker
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
GRPC_HOST=localhost GRPC_PORT=50051 GRPC_WORKER_API_KEY=your-key \
OPENAI_API_KEY=sk-... python -m worker.main
```

---

## Configuration Reference

All configuration is loaded from `.env` (file) then overridden by environment variables.

### Server

| Env var | Default | Description |
|---|---|---|
| `SERVER_HOST` | `0.0.0.0` | HTTP bind address |
| `SERVER_PORT` | `8080` | HTTP port |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000` | Comma-separated allowed origins (`*` for all) |

### Database (PostgreSQL)

| Env var | Default | Description |
|---|---|---|
| `DB_HOST` | `localhost` | Host |
| `DB_PORT` | `5433` | Port |
| `DB_USER` | `aiox` | Username |
| `DB_PASSWORD` | — | **Required** |
| `DB_NAME` | `aiox` | Database name |
| `DB_SSLMODE` | `disable` | `disable` / `require` / `verify-full` |
| `DB_MAX_CONNS` | `25` | Connection pool max |
| `DB_MIN_CONNS` | `2` | Connection pool min |
| `DB_AUTO_MIGRATE` | `false` | Run migrations on startup |
| `DB_MIGRATIONS_PATH` | `./migrations` | Path to SQL migrations |

### Redis

| Env var | Default | Description |
|---|---|---|
| `REDIS_HOST` | `localhost` | Host |
| `REDIS_PORT` | `6379` | Port |
| `REDIS_PASSWORD` | — | Optional password |
| `REDIS_DB` | `0` | Database index |

### JWT

| Env var | Default | Description |
|---|---|---|
| `JWT_ACCESS_SECRET` | — | **Required**, ≥32 chars, must differ from refresh |
| `JWT_REFRESH_SECRET` | — | **Required**, ≥32 chars |
| `JWT_ACCESS_EXPIRY` | `15m` | Access token lifetime |
| `JWT_REFRESH_EXPIRY` | `168h` | Refresh token lifetime (7 days) |

### Encryption

| Env var | Default | Description |
|---|---|---|
| `ENCRYPTION_KEY` | — | **Required** — 64 hex chars (32-byte AES-256 key) |

Generate a key:
```bash
openssl rand -hex 32
```

### XMPP

| Env var | Default | Description |
|---|---|---|
| `XMPP_DOMAIN` | `aiox.local` | XMPP domain |
| `XMPP_COMPONENT_HOST` | `localhost` | ejabberd host |
| `XMPP_COMPONENT_PORT` | `5275` | ejabberd component port |
| `XMPP_COMPONENT_SECRET` | `component_secret` | Shared secret (matches ejabberd.yml) |
| `XMPP_COMPONENT_NAME` | `agents.aiox.local` | Component subdomain |

### NATS

| Env var | Default | Description |
|---|---|---|
| `NATS_URL` | `nats://localhost:4222` | NATS connection URL |

### gRPC (Worker)

| Env var | Default | Description |
|---|---|---|
| `GRPC_HOST` | `0.0.0.0` | gRPC bind address |
| `GRPC_PORT` | `50051` | gRPC port |
| `GRPC_WORKER_API_KEY` | — | **Required**, ≥32 chars |
| `GRPC_TASK_TIMEOUT_SEC` | `120` | Max task execution time |

### Governance

| Env var | Default | Description |
|---|---|---|
| `GOVERNANCE_MAX_TOKENS_PER_DAY` | `100000` | Token quota per user per day |
| `GOVERNANCE_MAX_TOKENS_PER_MINUTE` | `10000` | Token rate limit per user per minute |
| `GOVERNANCE_MAX_REQUESTS_PER_DAY` | `1000` | Request quota per user per day |

### Logging

| Env var | Default | Options |
|---|---|---|
| `LOG_LEVEL` | `debug` | `debug` `info` `warn` `error` |
| `LOG_FORMAT` | `text` | `text` `json` |

---

## REST API Reference

Base URL: `http://localhost:8080`

All protected endpoints require the header:
```
Authorization: Bearer <access_token>
```

### Health & Metrics

```
GET  /health/live         # Liveness probe — always 200
GET  /health/ready        # Readiness probe — checks DB + NATS + workers
GET  /metrics             # Prometheus metrics
```

---

### Authentication

#### Register

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "strongpassword"
}
```

Response `201`:
```json
{
  "id": "uuid",
  "email": "user@example.com",
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "strongpassword"
}
```

Response `200`:
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_in": 900
}
```

#### Refresh Token

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJ..."
}
```

#### Logout

```http
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

---

### Agents

#### Create Agent

```http
POST /api/v1/agents/
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "My Assistant",
  "description": "A helpful assistant",
  "system_prompt": "You are a helpful assistant. Be concise.",
  "llm_config": {
    "provider": "openai",
    "model": "gpt-4o-mini",
    "temperature": 0.7,
    "max_tokens": 1024
  },
  "memory_config": {
    "short_term_limit": 20,
    "long_term_enabled": true,
    "long_term_limit": 10
  },
  "governance": {
    "blocked": false,
    "allowed_providers": ["openai", "anthropic"],
    "allowed_domains": []
  }
}
```

Response `201`:
```json
{
  "id": "uuid",
  "name": "My Assistant",
  "jid": "agent-uuid@agents.aiox.local",
  "owner_user_id": "user-uuid",
  "created_at": "2024-01-01T00:00:00Z"
}
```

#### List Agents

```http
GET /api/v1/agents/
Authorization: Bearer <access_token>
```

#### Get Agent

```http
GET /api/v1/agents/{agentID}
Authorization: Bearer <access_token>
```

#### Update Agent

```http
PUT /api/v1/agents/{agentID}
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Updated Name",
  "system_prompt": "New system prompt",
  "llm_config": { "provider": "anthropic", "model": "claude-haiku-4-5-20251001" }
}
```

#### Delete Agent

```http
DELETE /api/v1/agents/{agentID}
Authorization: Bearer <access_token>
```

---

### Agent Memory

#### List Memories

```http
GET /api/v1/agents/{agentID}/memories/
Authorization: Bearer <access_token>
```

#### Create Memory

```http
POST /api/v1/agents/{agentID}/memories/
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "content": "User prefers concise answers",
  "memory_type": "preference"
}
```

#### Semantic Search

```http
POST /api/v1/agents/{agentID}/memories/search
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "query": "user preferences",
  "limit": 5
}
```

#### Delete Single Memory

```http
DELETE /api/v1/agents/{agentID}/memories/{memoryID}
Authorization: Bearer <access_token>
```

#### Delete All Memories

```http
DELETE /api/v1/agents/{agentID}/memories/
Authorization: Bearer <access_token>
```

---

### Governance

#### Get Quota

```http
GET /api/v1/governance/quota
Authorization: Bearer <access_token>
```

Response:
```json
{
  "tokens_used_today": 1234,
  "tokens_limit_day": 100000,
  "requests_today": 10,
  "requests_limit_day": 1000
}
```

#### Audit Logs (all agents)

```http
GET /api/v1/governance/audit?limit=50&offset=0&severity=warn
Authorization: Bearer <access_token>
```

#### Audit Logs (single agent)

```http
GET /api/v1/agents/{agentID}/audit?limit=20
Authorization: Bearer <access_token>
```

---

### LLM Providers and Models

| Provider | `provider` value | Example models |
|---|---|---|
| OpenAI | `openai` | `gpt-4o`, `gpt-4o-mini`, `o1-mini` |
| Anthropic | `anthropic` | `claude-sonnet-4-6`, `claude-haiku-4-5-20251001` |
| Ollama (local) | `ollama` | `llama3.2`, `mistral`, `phi3` |

---

## Using XMPP to Chat with Agents

Once you have:
1. An XMPP account (`wande@aiox.local`)
2. An agent created via REST API (note its `jid`, e.g. `agent-uuid@agents.aiox.local`)
3. Connected with your XMPP client

**Add the agent as a contact** in your XMPP client using its JID:
```
agent-uuid@agents.aiox.local
```

**Send a message** — the platform will:
1. Receive the XMPP stanza via ejabberd → XMPP Component
2. Publish it to NATS (`aiox.messages.inbound`)
3. Orchestrator validates ownership + governance quotas
4. Dispatcher fetches agent config + memory context
5. Sends the task to the Python worker via gRPC
6. Worker calls the configured LLM
7. Response flows back: gRPC → NATS outbound → XMPP Component → ejabberd → your client

---

## Python Worker

The Python worker connects to the Go API via gRPC and processes LLM tasks.

### Environment Variables

| Env var | Default | Description |
|---|---|---|
| `WORKER_ID` | `worker-{pid}` | Unique identifier |
| `GRPC_HOST` | `localhost` | API server hostname |
| `GRPC_PORT` | `50051` | gRPC port |
| `GRPC_WORKER_API_KEY` | — | Must match `GRPC_WORKER_API_KEY` in API config |
| `MAX_CONCURRENT` | `4` | Max parallel tasks |
| `OPENAI_API_KEY` | — | Enables OpenAI provider |
| `ANTHROPIC_API_KEY` | — | Enables Anthropic provider |
| `OLLAMA_BASE_URL` | `http://localhost:11434` | Ollama endpoint (always enabled) |

### Running multiple workers

```bash
# Worker 1 — OpenAI
WORKER_ID=worker-openai OPENAI_API_KEY=sk-... python -m worker.main &

# Worker 2 — Anthropic
WORKER_ID=worker-anthropic ANTHROPIC_API_KEY=sk-ant-... python -m worker.main &
```

The Go API's **worker pool** automatically distributes tasks using least-loaded selection.

---

## Make Targets

```bash
make build           # Compile to ./bin/aiox-api
make dev             # Run API with go run (hot-reload friendly)
make up              # docker compose up -d (all services)
make down            # docker compose down
make docker-build    # Build Go API Docker image

make test            # Unit tests (no Docker needed)
make test-integration # Integration tests (requires Docker)
make test-coverage   # Coverage report → coverage.html

make migrate-up      # Apply all pending DB migrations
make migrate-create  # Create a new migration (prompts for name)

make vet             # go vet ./...
make fmt             # gofmt -w .
make fmt-check       # Verify formatting (CI-safe)
make lint            # golangci-lint
make security        # govulncheck
make check           # fmt-check + vet + test

make proto           # Regenerate gRPC code from worker.proto
```

---

## Running Tests

### Unit tests (no infrastructure required)

```bash
make test
# or
go test ./internal/... -v -race -count=1
```

### Single test

```bash
go test ./internal/config/ -run TestValidate_ValidConfig -v
```

### Integration tests (requires Docker)

Integration tests use [testcontainers](https://testcontainers.com/) to spin up PostgreSQL and Redis automatically.

```bash
make test-integration
# or
go test ./tests/... -v -race -count=1 -tags=integration
```

### Test coverage

```bash
make test-coverage
# Opens coverage.html in your browser
```

---

## Project Structure

```
aiox/
├── cmd/api/main.go              # Entry point: wires all services
├── internal/
│   ├── api/                     # HTTP router, response helpers
│   ├── auth/                    # JWT, bcrypt, AES-256-GCM
│   ├── agents/                  # Agent CRUD + ownership middleware
│   ├── config/                  # Koanf config + validation
│   ├── database/                # pgxpool + auto-migration
│   ├── redis/                   # Redis client
│   ├── nats/                    # JetStream client, publisher, consumer
│   ├── xmpp/                    # XMPP component, handler, outbound relay
│   ├── orchestrator/            # Event loop, router, validator
│   ├── worker/                  # gRPC server, pool, dispatcher, auth
│   ├── memory/                  # Short-term (Redis) + long-term (pgvector)
│   ├── governance/              # Quota, rate limiting, audit logs
│   ├── middleware/              # Logging, CORS, security headers, metrics
│   ├── metrics/                 # Prometheus metric definitions
│   └── users/                   # User model + repository
├── worker/                      # Python AI worker
│   ├── Dockerfile
│   ├── requirements.txt
│   └── worker/
│       ├── main.py              # Entry point
│       ├── config.py            # Env var config
│       ├── client.py            # gRPC client loop
│       ├── embedding.py         # sentence-transformers
│       ├── memory.py            # Memory context builder
│       └── llm/                 # OpenAI, Anthropic, Ollama providers
├── proto/worker/v1/worker.proto # gRPC service definition
├── migrations/                  # 10 SQL migrations (golang-migrate)
├── tests/integration/           # Integration test suite
├── docker/
│   ├── ejabberd/
│   │   ├── ejabberd.yml         # XMPP server config
│   │   ├── gen-cert.sh          # Generate self-signed TLS cert
│   │   └── install-ca.sh        # Install CA in Ubuntu trust store
│   └── postgres/init.sql        # DB initialization
├── .github/workflows/ci.yml     # GitHub Actions: test + lint + build
├── Dockerfile                   # Multi-stage Go API image
├── docker-compose.yml           # Full stack
├── Makefile
└── .env.example                 # Configuration template
```

---

## Troubleshooting

### Dino: "Cannot establish a secure connection"

The ejabberd TLS certificate is not trusted by your OS. Run:

```bash
bash docker/ejabberd/gen-cert.sh
sudo bash docker/ejabberd/install-ca.sh
docker compose restart ejabberd
```

Then restart Dino completely.

### API fails to start: "jwt access secret must be at least 32 characters"

Your `.env` has short or missing JWT secrets. Generate safe values:

```bash
openssl rand -base64 48   # use output as JWT_ACCESS_SECRET
openssl rand -base64 48   # use output as JWT_REFRESH_SECRET
openssl rand -hex 32      # use output as ENCRYPTION_KEY
openssl rand -base64 48   # use output as GRPC_WORKER_API_KEY
```

### No workers connected (`"workers": 0` in /health/ready)

- Check the worker is running: `docker compose logs aiox-worker`
- Verify `GRPC_WORKER_API_KEY` matches between API and worker
- Confirm the worker can reach the API on port 50051

### ejabberd component connection refused

- Check `XMPP_COMPONENT_SECRET` matches the password in `docker/ejabberd/ejabberd.yml`
- Verify ejabberd is healthy: `docker compose ps ejabberd`
- Check port 5275 is accessible: `nc -zv localhost 5275`

### Messages not reaching the worker

Check the NATS monitor to inspect stream state:
```
http://localhost:8222/jsz?streams=true&consumers=true
```

Check orchestrator and dispatcher logs:
```bash
docker compose logs aiox-api | grep -E "orchestrator|dispatcher|error"
```

### go vet warning in internal/xmpp/component.go

This is a pre-existing upstream issue in `gosrc.io/xmpp` (lock copy). It does not affect functionality and is not our bug.

---

## License

MIT
