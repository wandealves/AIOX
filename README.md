# AIOX — Multi-Agent AI Platform

AIOX is a production-ready multi-agent AI platform built in Go. Users interact with AI agents over **XMPP**, **REST API**, or a **real-time WebSocket chat**; agents run in isolated **Python workers** connected via **gRPC**, with **NATS JetStream** as the async message bus, **PostgreSQL** for persistence and vector memory, and **Redis** for caching and rate limiting. A **Next.js dashboard** provides a web UI for agent management, conversations, and governance monitoring.

Agents support **tool/function calling** (HTTP API invocation), **pipeline chaining** (sequential multi-agent workflows), **cron scheduling**, and **multimodal file attachments**. The platform includes **OpenTelemetry distributed tracing** (Jaeger), **Prometheus metrics** with **Grafana dashboards**, a **dead-letter queue** for failed messages, and a **circuit breaker** on worker connections.

---

## Table of Contents

- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Quick Start (Docker)](#quick-start-docker)
- [Frontend Dashboard](#frontend-dashboard)
- [TLS Setup for XMPP Clients](#tls-setup-for-xmpp-clients)
- [Registering XMPP Users](#registering-xmpp-users)
- [Local Development](#local-development)
- [Configuration Reference](#configuration-reference)
- [REST API Reference](#rest-api-reference)
- [WebSocket Chat](#websocket-chat)
- [Organizations (Multi-Tenancy)](#organizations-multi-tenancy)
- [Agent Tools (Function Calling)](#agent-tools-function-calling)
- [Pipelines (Agent Chaining)](#pipelines-agent-chaining)
- [Scheduled Tasks](#scheduled-tasks)
- [File Attachments](#file-attachments)
- [Observability & Monitoring](#observability--monitoring)
- [Using XMPP to Chat with Agents](#using-xmpp-to-chat-with-agents)
- [Python Worker](#python-worker)
- [Make Targets](#make-targets)
- [Running Tests](#running-tests)
- [Project Structure](#project-structure)
- [Troubleshooting](#troubleshooting)

---

## Architecture

```
                  ┌──────────────────────────────────────────────────────────┐
                  │                      NATS JetStream                      │
                  │  inbound ──► Orchestrator ──► tasks ──► Dispatcher       │
                  │              (validate·route·quota)   (agent+memory+tools│
                  │                                        +circuit breaker) │
                  └──────┬──────────────────────────────────────┬────────────┘
                         │                                      │
                         ▲                                      ▼
                         │                        gRPC ──► Python Worker ──► Tool HTTP calls
                         │                        (OpenAI · Anthropic · Ollama)
                         │                                      │
                         │                       ┌──────────────┤
                         │                 Pipeline?       NATS (outbound)
                         │                  Next step        ┌────┴────┐
Scheduler (cron) ────────┤                  via NATS         │         │
                         │                                   │         │
User ──XMPP──► ejabberd ──► XMPP Component                  │         │
                                  ▲                          │         │
                                  └── XMPP Relay ◄───────────┘         │
                                                                       │
Browser ──► Next.js Dashboard ──► REST API                             │
                   │                  │                                 │
                   │             POST /messages ──► NATS (inbound)      │
                   │             POST /attachments (multipart upload)   │
                   │                                                   │
                   └──── WebSocket ◄──── WS Relay ◄────────────────────┘

                         Failed messages (>5 retries) ──► DLQ Stream
                         All requests ──► Jaeger (traces) + Prometheus (metrics) + Grafana
```

### Stack

| Layer                 | Technology                      |
| --------------------- | ------------------------------- |
| Frontend              | Next.js 14 + TypeScript + Tailwind |
| HTTP API              | Go + chi                        |
| Real-time             | WebSocket (nhooyr.io/websocket) |
| Async messaging       | NATS JetStream (4 streams + DLQ)|
| XMPP server           | ejabberd                        |
| AI workers            | Python 3.12 + gRPC              |
| Database              | PostgreSQL 16 + pgvector        |
| Cache / Rate limiting | Redis 7                         |
| Tracing               | OpenTelemetry + Jaeger          |
| Metrics               | Prometheus + Grafana            |
| Scheduling            | robfig/cron (Go)                |

---

## Prerequisites

| Tool           | Minimum version | Purpose                                         |
| -------------- | --------------- | ----------------------------------------------- |
| Docker         | 24+             | All services                                    |
| Docker Compose | v2              | Orchestration                                   |
| Go             | 1.24            | Local dev / unit tests                          |
| Node.js        | 20+             | Frontend local dev                              |
| Python         | 3.12            | Worker local dev                                |
| `openssl`      | any             | TLS certificate generation                      |
| An XMPP client | —               | Chat with agents (e.g. [Dino](https://dino.im)) |

> **Go path note:** if Go is installed at `~/go-sdk/go/bin`, add it to your PATH:
>
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

| Service        | Port(s)          | Description                         |
| -------------- | ---------------- | ----------------------------------- |
| PostgreSQL     | 5433             | Primary database                    |
| Redis          | 6379             | Cache + rate limiting               |
| NATS           | 4222, 8222       | Message bus (HTTP monitor at :8222) |
| ejabberd       | 5222, 5275, 5280 | XMPP server                         |
| aiox-api       | 8080, 50051      | REST API + gRPC + WebSocket         |
| aiox-worker    | —                | Python AI worker                    |
| aiox-frontend  | 3000             | Next.js dashboard                   |
| Jaeger         | 16686, 4317      | Trace collector + UI                |
| Prometheus     | 9090             | Metrics scraper + alerting          |
| Grafana        | 3001             | Dashboards + visualization          |

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

## Frontend Dashboard

AIOX includes a **Next.js 14** web dashboard for managing agents, chatting in real-time, and monitoring governance.

### Pages

| Page | Path | Description |
| ---- | ---- | ----------- |
| Login | `/login` | Email + password authentication |
| Register | `/register` | New account creation |
| Dashboard | `/dashboard` | Overview with agent count, token usage, request stats |
| Agents | `/agents` | List, create, edit, delete agents |
| Agent Chat | `/agents/{id}/chat` | Real-time chat via WebSocket |
| Quotas | `/governance/quota` | Token and request usage visualization |
| Audit Logs | `/governance/audit` | Filterable audit log viewer |
| Admin Users | `/admin/users` | User management (admin only) |
| Admin Stats | `/admin/stats` | Platform-wide statistics (admin only) |

### Running locally

```bash
cd frontend
npm install
npm run dev
# Dashboard available at http://localhost:3000
```

### Environment variables

Create `frontend/.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8080
```

### Chat flow

1. User opens `/agents/{id}/chat`
2. History loads via `GET /api/v1/agents/{id}/conversations`
3. WebSocket connects to `ws://host/api/v1/ws?token=<JWT>`
4. User sends a message via `POST /api/v1/agents/{id}/messages` (returns 202 + request_id)
5. Agent response arrives via WebSocket and renders in real-time

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
docker exec -it aiox-ejabberd ejabberdctl register turing aiox.local senha123

# List registered users
docker exec -it aiox-ejabberd ejabberdctl registered_users aiox.local

# Delete a user
docker exec -it aiox-ejabberd ejabberdctl unregister <username> aiox.local
```

### Connecting with Dino

1. Open Dino → **Add Account**
2. JID: `turing@aiox.local`
3. Password: `senha123`
4. Dino will auto-discover the server at `aiox.local:5222`
5. The certificate should now be trusted (after `install-ca.sh`)

---

## Local Development

### 1. Start infrastructure only

```bash
docker compose up -d postgres redis nats ejabberd

# Optional: include monitoring stack
docker compose up -d postgres redis nats ejabberd jaeger prometheus grafana
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

| Env var                | Default                 | Description                                   |
| ---------------------- | ----------------------- | --------------------------------------------- |
| `SERVER_HOST`          | `0.0.0.0`               | HTTP bind address                             |
| `SERVER_PORT`          | `8080`                  | HTTP port                                     |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000` | Comma-separated allowed origins (`*` for all) |

### Database (PostgreSQL)

| Env var              | Default        | Description                           |
| -------------------- | -------------- | ------------------------------------- |
| `DB_HOST`            | `localhost`    | Host                                  |
| `DB_PORT`            | `5433`         | Port                                  |
| `DB_USER`            | `aiox`         | Username                              |
| `DB_PASSWORD`        | —              | **Required**                          |
| `DB_NAME`            | `aiox`         | Database name                         |
| `DB_SSLMODE`         | `disable`      | `disable` / `require` / `verify-full` |
| `DB_MAX_CONNS`       | `25`           | Connection pool max                   |
| `DB_MIN_CONNS`       | `2`            | Connection pool min                   |
| `DB_AUTO_MIGRATE`    | `false`        | Run migrations on startup             |
| `DB_MIGRATIONS_PATH` | `./migrations` | Path to SQL migrations                |

### Redis

| Env var          | Default     | Description       |
| ---------------- | ----------- | ----------------- |
| `REDIS_HOST`     | `localhost` | Host              |
| `REDIS_PORT`     | `6379`      | Port              |
| `REDIS_PASSWORD` | —           | Optional password |
| `REDIS_DB`       | `0`         | Database index    |

### JWT

| Env var              | Default | Description                                       |
| -------------------- | ------- | ------------------------------------------------- |
| `JWT_ACCESS_SECRET`  | —       | **Required**, ≥32 chars, must differ from refresh |
| `JWT_REFRESH_SECRET` | —       | **Required**, ≥32 chars                           |
| `JWT_ACCESS_EXPIRY`  | `15m`   | Access token lifetime                             |
| `JWT_REFRESH_EXPIRY` | `168h`  | Refresh token lifetime (7 days)                   |

### Encryption

| Env var          | Default | Description                                       |
| ---------------- | ------- | ------------------------------------------------- |
| `ENCRYPTION_KEY` | —       | **Required** — 64 hex chars (32-byte AES-256 key) |

Generate a key:

```bash
openssl rand -hex 32
```

### XMPP

| Env var                 | Default             | Description                          |
| ----------------------- | ------------------- | ------------------------------------ |
| `XMPP_DOMAIN`           | `aiox.local`        | XMPP domain                          |
| `XMPP_COMPONENT_HOST`   | `localhost`         | ejabberd host                        |
| `XMPP_COMPONENT_PORT`   | `5275`              | ejabberd component port              |
| `XMPP_COMPONENT_SECRET` | `component_secret`  | Shared secret (matches ejabberd.yml) |
| `XMPP_COMPONENT_NAME`   | `agents.aiox.local` | Component subdomain                  |

### NATS

| Env var    | Default                 | Description         |
| ---------- | ----------------------- | ------------------- |
| `NATS_URL` | `nats://localhost:4222` | NATS connection URL |

### gRPC (Worker)

| Env var                 | Default   | Description             |
| ----------------------- | --------- | ----------------------- |
| `GRPC_HOST`             | `0.0.0.0` | gRPC bind address       |
| `GRPC_PORT`             | `50051`   | gRPC port               |
| `GRPC_WORKER_API_KEY`   | —         | **Required**, ≥32 chars |
| `GRPC_TASK_TIMEOUT_SEC` | `120`     | Max task execution time |

### Governance

| Env var                            | Default  | Description                          |
| ---------------------------------- | -------- | ------------------------------------ |
| `GOVERNANCE_MAX_TOKENS_PER_DAY`    | `100000` | Token quota per user per day         |
| `GOVERNANCE_MAX_TOKENS_PER_MINUTE` | `10000`  | Token rate limit per user per minute |
| `GOVERNANCE_MAX_REQUESTS_PER_DAY`  | `1000`   | Request quota per user per day       |

### Tracing (OpenTelemetry)

| Env var                  | Default          | Description                            |
| ------------------------ | ---------------- | -------------------------------------- |
| `TRACING_ENABLED`        | `false`          | Enable distributed tracing             |
| `TRACING_OTLP_ENDPOINT`  | `localhost:4317` | OTLP gRPC collector (e.g. Jaeger)      |
| `TRACING_SERVICE_NAME`   | `aiox-api`       | Service name in traces                 |
| `TRACING_SAMPLE_RATE`    | `1.0`            | Sampling rate (0.0–1.0)               |

### Storage (Attachments)

| Env var               | Default              | Description                      |
| --------------------- | -------------------- | -------------------------------- |
| `STORAGE_BACKEND`     | `local`              | Storage backend (`local`)        |
| `STORAGE_LOCAL_PATH`  | `./data/attachments` | Local filesystem path            |
| `STORAGE_MAX_SIZE_MB` | `10`                 | Max upload size in MB            |

### Logging

| Env var      | Default | Options                       |
| ------------ | ------- | ----------------------------- |
| `LOG_LEVEL`  | `debug` | `debug` `info` `warn` `error` |
| `LOG_FORMAT` | `text`  | `text` `json`                 |

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
    "enabled": true,
    "max_short_term_msgs": 20
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

### Chat (REST + WebSocket)

#### Send Message to Agent

```http
POST /api/v1/agents/{agentID}/messages
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "body": "Hello, what can you help me with?"
}
```

Response `202`:

```json
{
  "data": {
    "request_id": "ws-uuid-20240101120000.000",
    "status": "accepted"
  }
}
```

The agent's response will arrive via WebSocket (see below).

#### List Conversations

```http
GET /api/v1/agents/{agentID}/conversations?page=1&page_size=20
Authorization: Bearer <access_token>
```

---

### WebSocket Chat

Connect to the WebSocket endpoint for real-time agent responses:

```
ws://localhost:8080/api/v1/ws?token=<JWT_access_token>
```

**Receiving messages** (server → client):

```json
{
  "type": "message",
  "agent_id": "uuid",
  "body": "Here's what I can help you with...",
  "request_id": "ws-uuid-20240101120000.000"
}
```

**Sending messages** (client → server):

```json
{
  "type": "message",
  "agent_id": "uuid",
  "body": "Tell me more"
}
```

---

### Organizations (Multi-Tenancy)

Organizations allow grouping users and agents together with role-based access control.

**Roles** (highest to lowest): `owner` > `admin` > `editor` > `viewer`

#### Create Organization

```http
POST /api/v1/orgs
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "My Team",
  "slug": "my-team",
  "description": "Our AI workspace"
}
```

Response `201`: The creator is automatically added as `owner`.

#### List My Organizations

```http
GET /api/v1/orgs
Authorization: Bearer <access_token>
```

#### Organization Detail / Update / Delete

```http
GET    /api/v1/orgs/{orgID}            # Any member
PUT    /api/v1/orgs/{orgID}            # Admin+
DELETE /api/v1/orgs/{orgID}            # Owner only
```

#### Members

```http
GET    /api/v1/orgs/{orgID}/members              # Any member
PUT    /api/v1/orgs/{orgID}/members/{userID}      # Admin+ (change role)
DELETE /api/v1/orgs/{orgID}/members/{userID}      # Admin+ (remove member)
```

Role update body:

```json
{ "role": "editor" }
```

#### Invites

```http
POST   /api/v1/orgs/{orgID}/invites        # Admin+ (create invite)
GET    /api/v1/orgs/{orgID}/invites         # Admin+ (list invites)
DELETE /api/v1/orgs/{orgID}/invites/{id}    # Admin+ (cancel invite)
POST   /api/v1/invites/{token}/accept       # Any authenticated user
```

Create invite body:

```json
{
  "email": "colleague@example.com",
  "role": "editor"
}
```

Invites expire after 7 days.

#### Organization Agents

```http
GET  /api/v1/orgs/{orgID}/agents            # Any member
POST /api/v1/orgs/{orgID}/agents            # Editor+
```

---

---

### Agent Tools (Function Calling)

Agents can invoke external HTTP APIs as tools during LLM execution. The LLM decides when to call tools based on the conversation context, and the Python worker executes the HTTP calls.

#### Create Tool

```http
POST /api/v1/agents/{agentID}/tools
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "get_weather",
  "description": "Get current weather for a city",
  "parameters": {
    "type": "object",
    "properties": {
      "city": { "type": "string", "description": "City name" }
    },
    "required": ["city"]
  },
  "endpoint_url": "https://api.weather.com/v1/current",
  "http_method": "GET",
  "auth_type": "api_key",
  "auth_config": { "header": "X-API-Key", "value": "your-key" },
  "timeout_sec": 10
}
```

#### List / Get / Update / Delete Tools

```http
GET    /api/v1/agents/{agentID}/tools              # List all tools
GET    /api/v1/agents/{agentID}/tools/{toolID}      # Get tool details
PUT    /api/v1/agents/{agentID}/tools/{toolID}      # Update tool
DELETE /api/v1/agents/{agentID}/tools/{toolID}      # Delete tool
```

Auth types: `""` (none), `"bearer"`, `"api_key"`. Auth config is encrypted at rest with AES-256-GCM.

The LLM tool-calling loop runs up to 10 iterations per request. Both OpenAI (`tool_calls` / `function`) and Anthropic (`tool_use`) formats are supported natively.

---

### Pipelines (Agent Chaining)

Pipelines execute a sequence of agents where each step's output feeds as input to the next. Go `text/template` transforms can modify the output between steps.

#### Create Pipeline

```http
POST /api/v1/pipelines
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Research & Summarize",
  "description": "Research a topic then summarize the findings",
  "steps": [
    {
      "agent_id": "research-agent-uuid",
      "timeout_sec": 120
    },
    {
      "agent_id": "summarizer-agent-uuid",
      "transform": "Summarize the following research:\n\n{{.Input}}",
      "timeout_sec": 60
    }
  ]
}
```

#### Execute Pipeline

```http
POST /api/v1/pipelines/{pipelineID}/execute
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "input": "What are the latest advances in quantum computing?"
}
```

Response `202`:

```json
{
  "id": "execution-uuid",
  "pipeline_id": "pipeline-uuid",
  "status": "running",
  "current_step": 0,
  "total_steps": 2
}
```

#### Other Pipeline Endpoints

```http
GET    /api/v1/pipelines                                          # List pipelines
GET    /api/v1/pipelines/{pipelineID}                              # Get pipeline
PUT    /api/v1/pipelines/{pipelineID}                              # Update pipeline
DELETE /api/v1/pipelines/{pipelineID}                              # Delete pipeline
GET    /api/v1/pipelines/{pipelineID}/executions                   # List executions
GET    /api/v1/pipelines/{pipelineID}/executions/{executionID}     # Get execution details
```

---

### Scheduled Tasks

Cron-based execution of agents or pipelines on a recurring schedule.

#### Create Scheduled Task

```http
POST /api/v1/schedules
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "Daily Report",
  "agent_id": "agent-uuid",
  "cron_expression": "0 9 * * *",
  "input_template": "Generate the daily report for today",
  "timezone": "America/Sao_Paulo"
}
```

Standard 5-field cron expressions (`minute hour day-of-month month day-of-week`). The scheduler polls every 60 seconds for due tasks.

#### Other Schedule Endpoints

```http
GET    /api/v1/schedules                    # List schedules
GET    /api/v1/schedules/{scheduleID}        # Get schedule
PUT    /api/v1/schedules/{scheduleID}        # Update schedule
DELETE /api/v1/schedules/{scheduleID}        # Delete schedule
```

---

### File Attachments

Upload files to send as multimodal input to agents. Supports images (JPEG, PNG, GIF, WebP), documents (PDF, plain text, CSV), and JSON.

#### Upload File

```http
POST /api/v1/attachments
Authorization: Bearer <access_token>
Content-Type: multipart/form-data

file=@photo.png
```

Response `201`:

```json
{
  "id": "attachment-uuid",
  "filename": "photo.png",
  "content_type": "image/png",
  "size_bytes": 45230,
  "checksum": "sha256:abc123..."
}
```

#### Other Attachment Endpoints

```http
GET    /api/v1/attachments                              # List uploads
GET    /api/v1/attachments/{attachmentID}                # Get metadata
GET    /api/v1/attachments/{attachmentID}/download       # Download file
DELETE /api/v1/attachments/{attachmentID}                # Delete file
```

**Max upload size:** 10 MB (configurable via `STORAGE_MAX_SIZE_MB`).

Images are sent to LLM providers as base64-encoded multimodal content (OpenAI `image_url`, Anthropic `image` with `base64` source). Text files are inlined as text content parts.

---

### Observability & Monitoring

AIOX includes a full observability stack for production monitoring.

#### Distributed Tracing (OpenTelemetry)

When `TRACING_ENABLED=true`, the platform emits traces via OTLP gRPC to Jaeger (or any OTLP-compatible collector). Traces span the full request lifecycle:

```
HTTP Request → Orchestrator → NATS → Dispatcher → gRPC → Python Worker → Tool calls
```

- W3C TraceContext propagated through NATS message headers
- `X-Trace-ID` response header on all HTTP requests
- `trace_id` and `span_id` in structured logs

Access the Jaeger UI at `http://localhost:16686`.

#### Prometheus Metrics

Available at `GET /metrics`. Key metrics:

| Metric | Type | Description |
| ------ | ---- | ----------- |
| `aiox_http_requests_total` | Counter | HTTP requests by method, path, status |
| `aiox_http_request_duration_seconds` | Histogram | Request latency |
| `aiox_tasks_dispatched_total` | Counter | Tasks sent to workers |
| `aiox_tasks_completed_total` | Counter | Tasks completed |
| `aiox_worker_pool_connected` | Gauge | Connected workers |
| `aiox_circuit_breaker_state` | Gauge | 0=closed, 1=open, 2=half-open |
| `aiox_circuit_breaker_trips_total` | Counter | Circuit breaker trips |
| `aiox_dlq_messages_total` | Counter | Dead-letter queue messages |

#### Grafana Dashboards

Pre-provisioned at `http://localhost:3001` with:
- HTTP request rates and latency
- Task throughput and worker pool status
- Circuit breaker state
- DLQ message accumulation

#### Alert Rules (Prometheus)

| Alert | Condition |
| ----- | --------- |
| `NoWorkersConnected` | `aiox_worker_pool_connected == 0` for 2 min |
| `HighErrorRate` | 5xx rate > 10% for 5 min |
| `CircuitBreakerOpen` | Circuit breaker open for 1 min |
| `DLQMessagesAccumulating` | > 10 DLQ messages/hour |

#### Dead-Letter Queue

Failed NATS messages (after 5 delivery attempts) are routed to the `AIOX_DLQ` stream with failure metadata. DLQ subjects: `aiox.dlq.tasks`, `aiox.dlq.messages`, `aiox.dlq.events`. Messages retained for 30 days.

#### Circuit Breaker

The gRPC dispatcher uses a circuit breaker to protect against worker failures:

- **Closed** (normal): requests pass through
- **Open** (after 5 consecutive failures): requests rejected, tasks Nak'd back to NATS
- **Half-Open** (after 30s timeout): one probe request allowed; success → Closed, failure → Open

---

### LLM Providers and Models

| Provider       | `provider` value | Example models                                   |
| -------------- | ---------------- | ------------------------------------------------ |
| OpenAI         | `openai`         | `gpt-4o`, `gpt-4o-mini`, `o1-mini`               |
| Anthropic      | `anthropic`      | `claude-sonnet-4-6`, `claude-haiku-4-5-20251001` |
| Ollama (local) | `ollama`         | `llama3.2`, `mistral`, `phi3`                    |

---

## Using XMPP to Chat with Agents

Once you have:

1. An XMPP account (`turing@aiox.local`)
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

The Python worker connects to the Go API via gRPC and processes LLM tasks. It supports tool/function calling (HTTP API execution), multimodal attachments (images + documents), and OpenTelemetry tracing.

### Environment Variables

| Env var               | Default                  | Description                                    |
| --------------------- | ------------------------ | ---------------------------------------------- |
| `WORKER_ID`           | `worker-{pid}`           | Unique identifier                              |
| `GRPC_HOST`           | `localhost`              | API server hostname                            |
| `GRPC_PORT`           | `50051`                  | gRPC port                                      |
| `GRPC_WORKER_API_KEY` | —                        | Must match `GRPC_WORKER_API_KEY` in API config |
| `MAX_CONCURRENT`      | `4`                      | Max parallel tasks                             |
| `OPENAI_API_KEY`      | —                        | Enables OpenAI provider                        |
| `ANTHROPIC_API_KEY`   | —                        | Enables Anthropic provider                     |
| `OLLAMA_BASE_URL`     | `http://localhost:11434` | Ollama endpoint (always enabled)               |
| `TRACING_ENABLED`     | `false`                  | Enable OpenTelemetry tracing                   |
| `TRACING_OTLP_ENDPOINT` | `localhost:4317`      | OTLP gRPC collector endpoint                   |

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
make build              # Compile to ./bin/aiox-api
make dev                # Run API with go run (hot-reload friendly)
make up                 # docker compose up -d (all services)
make down               # docker compose down
make docker-build       # Build Go API Docker image

make test               # Unit tests (no Docker needed)
make test-integration   # Integration tests (requires Docker)
make test-coverage      # Coverage report → coverage.html

make migrate-up         # Apply all pending DB migrations
make migrate-create     # Create a new migration (prompts for name)

make vet                # go vet ./...
make fmt                # gofmt -w .
make fmt-check          # Verify formatting (CI-safe)
make lint               # golangci-lint
make security           # govulncheck
make check              # fmt-check + vet + test

make proto              # Regenerate gRPC code from worker.proto

make frontend-install   # Install frontend dependencies
make frontend-dev       # Run frontend dev server (port 3000)
make frontend-build     # Build frontend for production
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
│   ├── chat/                    # REST chat (send message, list conversations)
│   ├── ws/                      # WebSocket hub, conn, handler, relay
│   ├── orgs/                    # Organizations (multi-tenancy)
│   ├── tools/                   # Agent tool definitions (function calling)
│   ├── pipelines/               # Agent chaining (sequential execution)
│   ├── scheduler/               # Cron-based task scheduling + runner
│   ├── attachments/             # File upload, storage, multimodal input
│   ├── tracing/                 # OpenTelemetry tracing + NATS propagation
│   ├── config/                  # Koanf config + validation
│   ├── database/                # pgxpool + auto-migration
│   ├── redis/                   # Redis client
│   ├── nats/                    # JetStream client, publisher, consumer, DLQ
│   ├── xmpp/                    # XMPP component, handler, outbound relay
│   ├── orchestrator/            # Event loop, router, validator
│   ├── worker/                  # gRPC server, pool, dispatcher, circuit breaker
│   ├── memory/                  # Short-term (Redis) + long-term (pgvector)
│   ├── governance/              # Quota, rate limiting, audit logs
│   ├── middleware/              # Logging, CORS, security headers, metrics
│   ├── metrics/                 # Prometheus metric definitions
│   └── users/                   # User model + repository
├── frontend/                    # Next.js 14 dashboard
│   ├── src/
│   │   ├── app/                 # App Router pages (14 routes)
│   │   ├── components/          # UI components (chat, agents, governance)
│   │   ├── hooks/               # React Query + WebSocket hooks
│   │   ├── lib/                 # API client, WS client, types
│   │   └── providers/           # Auth + Query providers
│   ├── Dockerfile               # Multi-stage Node.js image
│   └── .env.local.example       # Frontend env template
├── worker/                      # Python AI worker
│   ├── Dockerfile
│   ├── requirements.txt
│   └── worker/
│       ├── main.py              # Entry point
│       ├── config.py            # Env var config
│       ├── client.py            # gRPC client loop (tools + attachments)
│       ├── tools.py             # HTTP tool executor (aiohttp)
│       ├── embedding.py         # sentence-transformers
│       ├── memory.py            # Memory context builder
│       └── llm/                 # OpenAI, Anthropic, Ollama providers
├── proto/worker/v1/worker.proto # gRPC service definition
├── migrations/                  # 20 SQL migrations (golang-migrate)
├── tests/integration/           # Integration test suite
├── docker/
│   ├── ejabberd/
│   │   ├── ejabberd.yml         # XMPP server config
│   │   ├── gen-cert.sh          # Generate self-signed TLS cert
│   │   └── install-ca.sh        # Install CA in Ubuntu trust store
│   ├── prometheus/
│   │   ├── prometheus.yml       # Scrape config (aiox-api target)
│   │   └── alerts.yml           # Alert rules
│   ├── grafana/
│   │   ├── provisioning/        # Datasources (Prometheus + Jaeger)
│   │   └── dashboards/          # AIOX overview dashboard
│   └── postgres/init.sql        # DB initialization
├── .github/workflows/ci.yml     # GitHub Actions: test + lint + build
├── Dockerfile                   # Multi-stage Go API image
├── docker-compose.yml           # Full stack (API + frontend + monitoring)
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

### No traces in Jaeger

- Ensure `TRACING_ENABLED=true` in your `.env`
- Verify Jaeger is running: `docker compose ps jaeger`
- Check OTLP endpoint: `TRACING_OTLP_ENDPOINT=jaeger:4317` (Docker) or `localhost:4317` (local dev)

### Circuit breaker is open (tasks being rejected)

The circuit breaker opens after 5 consecutive gRPC failures. Check worker connectivity:

```bash
docker compose logs aiox-worker --tail=50
curl http://localhost:8080/metrics | grep circuit_breaker
```

It will auto-recover after 30 seconds (half-open state) if a probe request succeeds.

### Messages stuck in DLQ

Inspect DLQ messages via NATS monitor:

```
http://localhost:8222/jsz?streams=true&consumers=true
```

DLQ messages include the original subject, failure reason, and attempt count.

### go vet warning in internal/xmpp/component.go

This is a pre-existing upstream issue in `gosrc.io/xmpp` (lock copy). It does not affect functionality and is not our bug.

---

## License

MIT
