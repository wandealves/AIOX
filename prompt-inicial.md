Você é um engenheiro de software sênior especialista em:

* **Go** — Plataforma principal, alta concorrência, sistemas distribuídos
* **Python** — Workers especializados em IA/LLM
* **XMPP** (RFC 6120/6121/6122)
* **Arquitetura multi-tenant**
* **Orquestração determinística**
* **Plataformas de agentes IA**
* **Infraestrutura escalável**
* **Security & Governance**

Sua missão é implementar uma **plataforma multi-agente federável baseada na RFC AIOX**, com **arquitetura híbrida Go + Python**, onde usuários podem criar seus próprios agentes IA via interface web, com **governança e políticas de execução**.

---

# 🧠 VISÃO DO SISTEMA

Estamos construindo:

# **AIOX — Agent Identity & Orchestration over XMPP**

## 🔄 Arquitetura Híbrida

| Camada | Linguagem | Responsabilidade |
|--------|-----------|------------------|
| API Gateway | **Go** | HTTP/REST, JWT Auth, Rate Limit |
| Orchestrator | **Go** | Routing, Ownership, Policy, NATS |
| XMPP Component | **Go** | Protocolo, SASL, Message Delivery |
| Policy Engine | **Go** | Governance, Quota, Audit |
| **AI Worker** | **Python** | **LLM Inference, RAG, Tool Execution, Embeddings** |
| Web UI | TypeScript | Frontend React/Vue |

### Por que Híbrido?

```
┌─────────────────────────────────────────────────────────────┐
│                    GO (Plataforma)                          │
│  ✅ Alta concorrência (goroutines)                          │
│  ✅ Baixa latência de rede                                  │
│  ✅ Tipagem forte + compilação                              │
│  ✅ XMPP libraries maduras                                  │
│  ✅ Memory footprint baixo                                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼ (gRPC / NATS)
┌─────────────────────────────────────────────────────────────┐
│                  PYTHON (AI Worker)                         │
│  ✅ Ecossistema LLM completo (LangChain, LlamaIndex)        │
│  ✅ Libraries de embeddings (sentence-transformers)         │
│  ✅ pgvector integration madura                             │
│  ✅ Tool execution flexível                                 │
│  ✅ Prototipagem rápida de novos modelos                    │
└─────────────────────────────────────────────────────────────┘
```

---

# 🏗 ARQUITETURA MACRO

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Web UI (TypeScript)                          │
├─────────────────────────────────────────────────────────────────────┤
│                      API Gateway (Go)                                │
│                   JWT Auth + Rate Limit                              │
├─────────────────────────────────────────────────────────────────────┤
│                    Orchestrator Layer (Go)                           │
│           Ownership Validation + Policy Engine + NATS                │
├─────────────────────────────────────────────────────────────────────┤
│                   XMPP Cluster (ejabberd)                            │
│              External Component (Go) + SASL Auth                     │
├─────────────────────────────────────────────────────────────────────┤
│                         Event Bus (NATS)                             │
│              ┌──────────────────┴──────────────────┐                │
│              ▼                                     ▼                │
│     ┌─────────────────┐                   ┌─────────────────┐       │
│     │  AI Worker Pool │                   │  AI Worker Pool │       │
│     │    (Python)     │                   │    (Python)     │       │
│     │   LLM + RAG     │                   │   LLM + RAG     │       │
│     └─────────────────┘                   └─────────────────┘       │
├─────────────────────────────────────────────────────────────────────┤
│          Postgres + pgvector │ Redis │ Object Storage                │
└─────────────────────────────────────────────────────────────────────┘
```

---

# 🧩 MODELO DE AGENTE CRIADO PELO USUÁRIO

Cada agente criado via UI deve conter:

```json
{
  "agent_id": "uuid",
  "owner_user_id": "uuid",
  "jid": "agent-uuid@agents.domain.com",
  "profile": {
    "name": "Legal Assistant",
    "description": "Especialista em contratos",
    "system_prompt": "Você é um especialista jurídico...",
    "personality_traits": ["formal", "analítico"],
    "encrypted": true
  },
  "llm_config": {
    "provider": "openai",
    "model": "gpt-4-turbo",
    "temperature": 0.7,
    "max_tokens": 4096,
    "worker_pool": "python-ai-workers"
  },
  "capabilities": {
    "domain_tags": ["law", "contracts"],
    "priority_weight": 0.6
  },
  "memory_config": {
    "type": "hybrid",
    "short_term_limit": 10,
    "long_term_enabled": true,
    "vector_index_partition": "user_id_hash",
    "embedding_model": "sentence-transformers/all-MiniLM-L6-v2"
  },
  "tools": [
    {"name": "document_parser", "policy": "read_only", "worker": "python"},
    {"name": "search_api", "policy": "rate_limited", "worker": "python"}
  ],
  "governance": {
    "max_tokens_per_minute": 1000,
    "max_memory_mb": 512,
    "allowed_domains": ["internal.company.com"],
    "audit_log_enabled": true
  },
  "visibility": "private"
}
```

---

# 🔐 ISOLAMENTO MULTI-TENANT

## Regra Obrigatória de Ownership

```go
// Go - Orchestrator Layer
// Antes de qualquer processamento:
if message.FromUserID != agent.OwnerUserID {
    return Error("ACCESS_DENIED: Agent ownership mismatch")
}
```

## Filtro Vetorial Obrigatório

```python
# Python - AI Worker
# TODAS as queries de memória DEVEM incluir:
query = """
    SELECT content, embedding 
    FROM agent_memories 
    WHERE owner_user_id = $1 
      AND agent_id = $2 
      AND embedding <=> $3 < $4
    ORDER BY embedding <=> $3
    LIMIT $5
"""
# Nunca confiar apenas em agent_id
```

## Níveis de Isolamento

| Nível | O que é isolado | Como | Camada |
|-------|-----------------|------|--------|
| L1 | Agentes | owner_user_id FK | Go (DB) |
| L2 | Memória | Partition + WHERE clause | Python (RAG) |
| L3 | Tools | Policy engine por tool | Go + Python |
| L4 | Quotas | Rate limit por user_id | Go (API) |
| L5 | Criptografia | Chaves por tenant | Go (Auth) |
| L6 | Worker Context | Isolated per request | Python (Worker) |

Nenhum usuário pode:

* Ver agentes de outro
* Interagir com agentes de outro
* Acessar memória de outro
* Herdar ferramentas sem permissão
* Vazar contexto entre workers Python

---

# 🖥 UI PARA CRIAÇÃO DE AGENTE

Criar Web UI com:

## Tela 1 — Criar Agente

Campos:

* Nome
* Descrição
* System Prompt (criptografado em repouso)
* Tags de domínio
* **Configuração LLM** (provider, model, temperature)
* Ferramentas disponíveis (com políticas)
* Configuração de memória
* Peso de prioridade
* **Políticas de Governança**
  * Limite de tokens/minuto
  * Limite de memória
  * Domínios permitidos (para tools de rede)
  * Audit log (on/off)

## Tela 2 — Lista de Agentes

* Listar agentes do usuário
* Editar
* Excluir (soft delete com retenção)
* Ver métricas
* **Ver logs de auditoria**
* **Status dos workers Python**

## Tela 3 — Chat com Agente

* Interface tipo chat
* Histórico persistido
* Estado do agente (thinking/responding)
* **Indicador de quota restante**
* **Latência Go → Python**

## Tela 4 — Dashboard de Governança

* Uso de tokens por agente
* Violações de policy
* Custos estimados
* Alertas de limite
* **Health dos AI Workers**

---

# 🧠 MEMÓRIA DO AGENTE

## Implementação Híbrida

| Tipo | Storage | Linguagem | Use Case |
|------|---------|-----------|----------|
| Short-term | Redis | **Go** | Contexto da sessão, cache rápido |
| Long-term | Postgres + pgvector | **Python** | RAG, embeddings, recuperação semântica |
| Hybrid | Ambos | **Go + Python** | Produção |

### Go — Short-term Memory (Redis)

```go
type ShortTermMemory struct {
    redis *redis.Client
}

func (m *ShortTermMemory) GetSession(ctx context.Context, sessionID string) ([]Message, error) {
    // Go gerencia sessões ativas com baixa latência
}
```

### Python — Long-term Memory (pgvector)

```python
class LongTermMemory:
    def __init__(self, db_url: str):
        self.db = asyncpg.create_pool(db_url)
        self.embedder = SentenceTransformer('all-MiniLM-L6-v2')
    
    async def search(self, owner_user_id: str, agent_id: str, query: str, limit: int):
        # Gera embedding em Python
        embedding = self.embedder.encode(query)
        
        # Query COM filtro de ownership obrigatório
        results = await self.db.fetch("""
            SELECT content, embedding 
            FROM agent_memories 
            WHERE owner_user_id = $1 
              AND agent_id = $2 
            ORDER BY embedding <=> $3
            LIMIT $4
        """, owner_user_id, agent_id, embedding, limit)
        
        return results
```

---

# 🛠 SISTEMA DE TOOLS COM GOVERNANCE

## Arquitetura de Tools

```
┌─────────────────────────────────────────────────────────────┐
│                    Tool Registry (Go)                       │
│              Catalogo global de ferramentas                 │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼ (gRPC)
┌─────────────────────────────────────────────────────────────┐
│                  Tool Executor (Python)                     │
│         Execução segura em sandbox isolado                  │
└─────────────────────────────────────────────────────────────┘
```

## Tool Interface

```go
// Go - Tool Definition
type ToolDefinition struct {
    Name string
    Description string
    InputSchema json.RawMessage
    Policy ToolPolicy
    WorkerType string // "python" | "go" | "external"
}

type ToolPolicy struct {
    RateLimitPerMinute int
    AllowedDomains []string
    RequiresApproval bool
    AuditLog bool
    ReadOnly bool
    SandboxRequired bool
}
```

```python
# Python - Tool Execution
class ToolExecutor:
    async def execute(self, tool_name: str, input_ dict, context: ToolContext) -> ToolResult:
        # Valida policy antes de executar
        await self.policy_engine.validate(context)
        
        # Executa em sandbox se necessário
        if context.policy.sandbox_required:
            return await self.sandbox_execute(tool_name, input_data)
        
        # Executa tool específica
        handler = self.registry.get(tool_name)
        return await handler.run(input_data)
```

## Tool Registry

| Tool | Worker | Policy Default | Risco |
|------|--------|---------------|-------|
| document_parser | Python | read_only | Baixo |
| search_api | Python | rate_limited + allowed_domains | Médio |
| code_executor | Python | requires_approval + sandbox | Alto |
| email_sender | Python | requires_approval + audit | Alto |
| database_query | Python | read_only + allowed_tables | Crítico |
| http_request | Python | allowed_domains + timeout | Médio |

---

# 🧠 ORQUESTRADOR COM AGENTES PRIVADOS

## Fluxo de Mensagem Híbrido

```
1. Mensagem chega via XMPP
2. XMPP Component (Go) valida autenticação SASL/JWT
3. Orchestrator (Go) recebe evento via NATS
4. ✅ VALIDAÇÃO DE OWNERSHIP (Go - blocking)
5. ✅ VALIDAÇÃO DE GOVERNANCE (Go - quota, policy)
6. ✅ VALIDAÇÃO DE TOOLS (Go - permissões)
7. Busca ACD do agente (Go)
8. ⚠️ Modelo Matemático (Go - apenas para priorização de fila)
9. 📤 Dispatch para AI Worker (Go → Python via gRPC/NATS)
10. 🧠 Processamento LLM + RAG (Python)
11. 📥 Retorna resultado (Python → Go)
12. Atualiza memória curta (Go - Redis)
13. Atualiza memória longa (Python - pgvector)
14. Log de auditoria (Go)
15. Retorna resposta via XMPP (Go)
```

## Protocolo Go → Python

```go
// Go - Request para AI Worker
type AIWorkerRequest struct {
    RequestID string `json:"request_id"`
    AgentID string `json:"agent_id"`
    OwnerUserID string `json:"owner_user_id"`
    Message string `json:"message"`
    SessionContext []Message `json:"session_context"`
    Tools []ToolDefinition `json:"tools"`
    LLMConfig LLMConfig `json:"llm_config"`
    Governance GovernanceConfig `json:"governance"`
    TimeoutMS int `json:"timeout_ms"`
}

type AIWorkerResponse struct {
    RequestID string `json:"request_id"`
    Success bool `json:"success"`
    Response string `json:"response"`
    ToolsCalled []ToolCall `json:"tools_called"`
    TokensUsed TokenUsage `json:"tokens_used"`
    LatencyMS int `json:"latency_ms"`
    Error string `json:"error,omitempty"`
}
```

```python
# Python - AI Worker Handler
class AIWorkerHandler:
    async def process(self, request: AIWorkerRequest) -> AIWorkerResponse:
        # Valida ownership novamente (defense in depth)
        if not await self.verify_ownership(request.agent_id, request.owner_user_id):
            raise SecurityError("Ownership verification failed")
        
        # Recupera memória de longo prazo
        memory = await self.memory.search(
            owner_user_id=request.owner_user_id,
            agent_id=request.agent_id,
            query=request.message
        )
        
        # Constrói prompt com contexto
        prompt = self.build_prompt(request, memory)
        
        # Executa LLM
        response = await self.llm.generate(prompt, request.llm_config)
        
        # Executa tools se necessário
        if response.tool_calls:
            tool_results = await self.execute_tools(response.tool_calls, request.governance)
            response = await self.llm.generate_with_tools(prompt, tool_results)
        
        # Salva memória
        await self.memory.store(request.owner_user_id, request.agent_id, request.message, response.response)
        
        return AIWorkerResponse(...)
```

---

# 📊 MODELO MATEMÁTICO (GO LAYER)

## Nota Crítica: Modelo Matemático

```
⚠️ PARA AGENTES PRIVADOS (JID DIRETO):

O modelo de seleção A* = argmax S(Ai) é BYPASSADO.
O agente já foi selecionado pelo endereço XMPP.

O modelo matemático (Go) é utilizado APENAS para:
1. Priorização de fila (QoS)
2. Alocação de workers Python
3. Load balancing entre workers
4. Cenários futuros de broadcast/discovery

Elegibilidade (para priorização):
E(Ai, M) = similarity(M, Ci.domain_tags)

Score (para QoS):
S(Ai) = αR + βT + γP + δC

Onde:
R = Reputation (histórico do agente)
T = Tool availability
P = Priority weight (configurado pelo owner)
C = Current load (cluster health + Python worker availability)
```

---

# 📊 MODELO DE GOVERNANCE

## Policy Engine (Go)

```go
type PolicyEngine interface {
    ValidateMessage(ctx context.Context, msg Message) error
    ValidateToolCall(ctx context.Context, call ToolCall) error
    CheckQuota(ctx context.Context, userID string) error
    LogAudit(ctx context.Context, event AuditEvent) error
    CheckWorkerAvailability(ctx context.Context, pool string) bool
}

type QuotaConfig struct {
    TokensPerMinute int
    TokensPerDay int
    MaxConcurrentRequests int
    MaxMemoryMB int
    MaxPythonWorkerTimeMS int
}
```

## Python Worker Health Check

```go
// Go monitora saúde dos workers Python
type WorkerHealth struct {
    WorkerID string
    Status string // "healthy", "degraded", "unavailable"
    ActiveRequests int
    AvgLatencyMS int
    LastHeartbeat time.Time
    MemoryUsageMB int
}

// Orchestrator faz load balancing entre workers
func (o *Orchestrator) selectWorker(request *AIWorkerRequest) (*WorkerHealth, error) {
    workers := o.getHealthyWorkers()
    if len(workers) == 0 {
        return nil, Error("No available AI workers")
    }
    
    // Seleciona worker com menor carga
    return selectLeastLoaded(workers), nil
}
```

---

# 🗄 BANCO DE DADOS

## Tabelas Principais

```sql
-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE,
    password_hash TEXT,
    created_at TIMESTAMPTZ,
    quota_config JSONB
);

-- Agents
CREATE TABLE agents (
    id UUID PRIMARY KEY,
    owner_user_id UUID NOT NULL REFERENCES users(id),
    jid TEXT UNIQUE,
    profile JSONB,
    llm_config JSONB,
    capabilities JSONB,
    memory_config JSONB,
    governance JSONB,
    visibility TEXT DEFAULT 'private',
    created_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    
    INDEX idx_agents_owner (owner_user_id)
);

-- Agent Memory (partitioned)
CREATE TABLE agent_memories (
    id UUID PRIMARY KEY,
    owner_user_id UUID NOT NULL,
    agent_id UUID NOT NULL,
    content TEXT,
    embedding vector(1536),
    created_at TIMESTAMPTZ,
    
    PARTITION BY HASH (owner_user_id)
);

-- Agent Tools
CREATE TABLE agent_tools (
    id UUID PRIMARY KEY,
    agent_id UUID NOT NULL,
    tool_name TEXT,
    policy JSONB,
    worker_type TEXT DEFAULT 'python',
    created_at TIMESTAMPTZ
);

-- Executions (Audit Log)
CREATE TABLE executions (
    id UUID PRIMARY KEY,
    owner_user_id UUID NOT NULL,
    agent_id UUID NOT NULL,
    input TEXT,
    output TEXT,
    tokens_used INT,
    tools_called JSONB,
    worker_id TEXT,
    duration_ms INT,
    go_latency_ms INT,
    python_latency_ms INT,
    created_at TIMESTAMPTZ,
    
    INDEX idx_executions_owner (owner_user_id, created_at)
);

-- User Quotas
CREATE TABLE user_quotas (
    user_id UUID PRIMARY KEY,
    tokens_used_today INT,
    tokens_used_minute INT,
    last_reset TIMESTAMPTZ,
    violations JSONB
);

-- Audit Log
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    owner_user_id UUID NOT NULL,
    event_type TEXT,
    severity TEXT,
    details JSONB,
    created_at TIMESTAMPTZ
);

-- AI Worker Registry
CREATE TABLE ai_workers (
    id UUID PRIMARY KEY,
    worker_id TEXT UNIQUE,
    host TEXT,
    port INT,
    status TEXT DEFAULT 'healthy',
    last_heartbeat TIMESTAMPTZ,
    capabilities JSONB,
    created_at TIMESTAMPTZ
);
```

---

# 🔐 SEGURANÇA

## Obrigatório

| Controle | Implementação | Camada |
|----------|--------------|--------|
| API Auth | JWT com expiração curta (15min) | Go |
| XMPP Auth | SASL SCRAM-SHA-256 ou JWT via XEP-0386 | Go |
| Go ↔ Python | mTLS + JWT interno | Ambos |
| RBAC | Owner-only para agentes privados | Go |
| Rate Limit | Redis-based, por user_id | Go |
| Quota | Postgres + Redis counter | Go |
| Encryption | AES-256 para system_prompt em repouso | Go |
| TLS | Obrigatório para XMPP, API e gRPC | Ambos |
| Vector Security | WHERE owner_user_id em TODAS queries | Python |
| Sandbox | Isolated Python processes para tools | Python |

## Comunicação Go ↔ Python

```
┌─────────────────────────────────────────────────────────────┐
│  Opção 1: gRPC (Recomendado para produção)                  │
│  ✅ Tipagem forte (protobuf)                                │
│  ✅ Streaming bidirecional                                  │
│  ✅ mTLS nativo                                             │
│  ✅ Performance alta                                        │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  Opção 2: NATS (Recomendado para eventos)                   │
│  ✅ Pub/Sub nativo                                          │
│  ✅ Queue groups para load balancing                        │
│  ✅ Fácil scaling                                           │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│  Opção 3: HTTP/REST (Fallback)                              │
│  ✅ Simples                                                 │
│  ⚠️ Mais latência                                           │
└─────────────────────────────────────────────────────────────┘
```

### Configuração Recomendada

```yaml
# docker-compose.yml
communication:
  primary: grpc  # Go ↔ Python sync requests
  async: nats    # Events, logs, metrics
  fallback: http # Health checks
```

---

# 🧠 EXTENSÃO DA RFC AIOX

## Agent Ownership Metadata (ACD Extension)

```json
{
  "agent_id": "uuid",
  "owner_user_id": "uuid",
  "visibility": "private",
  "governance_version": "1.0",
  "policy_hash": "sha256:...",
  "worker_pool": "python-ai-workers",
  "llm_provider": "openai"
}
```

## XMPP Component Configuration

```xml
<!-- ejabberd.yml -->
components:
  "agents.domain.com":
    module: ejabberd_component
    host: "orchestrator.internal"
    port: 5280
    password: "shared_secret"
```

## gRPC Service Definition

```protobuf
// ai_worker.proto
service AIWorker {
    rpc ProcessMessage(AIWorkerRequest) returns (AIWorkerResponse);
    rpc StreamResponse(stream AIWorkerRequest) returns (stream AIWorkerResponse);
    rpc HealthCheck(HealthCheckRequest) returns (HealthCheckResponse);
}

message AIWorkerRequest {
    string request_id = 1;
    string agent_id = 2;
    string owner_user_id = 3;
    string message = 4;
    repeated Message session_context = 5;
    LLMConfig llm_config = 6;
    GovernanceConfig governance = 7;
}

message AIWorkerResponse {
    string request_id = 1;
    bool success = 2;
    string response = 3;
    TokenUsage tokens_used = 4;
    int32 latency_ms = 5;
    string error = 6;
}
```

---

# 📦 ESTRUTURA DO PROJETO

```
/
├── /cmd
│   ├── /api              # Go - API Gateway
│   ├── /orchestrator     # Go - Main Orchestrator
│   ├── /xmpp-component   # Go - XMPP External Component
│   ├── /policy-engine    # Go - Governance & Policy
│   └── /ai-worker        # Python - AI/LLM Worker
│
├── /internal
│   ├── /go
│   │   ├── /agents
│   │   ├── /useragents
│   │   ├── /acds
│   │   ├── /memory       # Short-term (Redis)
│   │   ├── /tools        # Tool definitions
│   │   ├── /scoring
│   │   ├── /xmpp
│   │   ├── /auth
│   │   ├── /storage
│   │   ├── /governance
│   │   └── /observability
│   │
│   └── /python
│       ├── /ai_worker
│       ├── /llm
│       ├── /rag
│       ├── /embeddings
│       ├── /tools        # Tool implementations
│       ├── /memory       # Long-term (pgvector)
│       └── /sandbox
│
├── /proto                # gRPC definitions
│   └── ai_worker.proto
│
├── /web
│   ├── /src
│   ├── /components
│   └── /pages
│
├── /docker
│   ├── /go-services
│   ├── /python-workers
│   └── /infrastructure
│
├── /tests
│   ├── /unit
│   ├── /integration
│   ├── /security
│   └── /e2e
│
├── /docs
│   └── rfc-aiox.md
│
├── go.mod
├── requirements.txt
└── docker-compose.yml
```

---

# 🐳 CONTAINERIZAÇÃO

## Serviços

| Serviço | Linguagem | Instâncias | Notas |
|---------|-----------|-----------|-------|
| api | **Go** | 3 | Load balanced |
| orchestrator | **Go** | 3 | Stateful, sticky sessions |
| xmpp-component | **Go** | 2 | ejabberd external component |
| **ai-worker** | **Python** | 3-10 | Auto-scale based on queue |
| policy-engine | **Go** | 2 | Stateless |
| ejabberd | Erlang | 3 | Cluster mode |
| nats | Go | 3 | Cluster mode |
| postgres | C | 1+1 | Primary + Replica |
| redis | C | 3 | Sentinel mode |
| web | Node | 2 | Static + CDN |

## Docker Compose (Excerpt)

```yaml
version: '3.8'

services:
  # Go Services
  api:
    build:
      context: .
      dockerfile: docker/go-services/api/Dockerfile
    deploy:
      replicas: 3
  
  orchestrator:
    build:
      context: .
      dockerfile: docker/go-services/orchestrator/Dockerfile
    deploy:
      replicas: 3
  
  xmpp-component:
    build:
      context: .
      dockerfile: docker/go-services/xmpp-component/Dockerfile
    deploy:
      replicas: 2
  
  # Python AI Workers
  ai-worker:
    build:
      context: .
      dockerfile: docker/python-workers/ai-worker/Dockerfile
    deploy:
      replicas: 3
      resources:
        limits:
          memory: 2G
        reservations:
          memory: 1G
    environment:
      - WORKER_POOL=python-ai-workers
      - GRPC_SERVER=orchestrator:50051
  
  # Infrastructure
  ejabberd:
    image: ejabberd/ecs
    deploy:
      replicas: 3
  
  nats:
    image: nats:latest
    command: ["-c", "/config/nats.conf"]
    deploy:
      replicas: 3
  
  postgres:
    image: pgvector/pgvector:pg16
    volumes:
      - postgres_/var/lib/postgresql/data
  
  redis:
    image: redis:7-alpine
    deploy:
      replicas: 3
```

---

# 🧪 TESTES

## Cobertura Mínima 80%

### Testes de Segurança

```go
// Go - Teste de isolamento multi-tenant
func TestMultiTenantIsolation(t *testing.T) {
    // User A não pode acessar agente de User B
}

// Go - Teste de violação de acesso
func TestOwnershipViolation(t *testing.T) {
    // Mensagem de user errado deve ser bloqueada
}

// Python - Teste de filtro vetorial
def test_vector_query_isolation():
    # Query de memória deve incluir owner_user_id
```

### Testes de Integração Go ↔ Python

```go
// Go - Teste de comunicação gRPC
func TestGoPythonGRPCCommunication(t *testing.T) {
    // Valida handshake, auth, e resposta
}

// Python - Teste de processamento LLM
def test_llm_processing():
    # Valida que LLM processa com contexto correto
}
```

### Testes Funcionais

```go
// Go
func TestAgentCreation(t *testing.T) {}
func TestAgentMemory(t *testing.T) {}
func TestScoringDeterminism(t *testing.T) {}
func TestPolicyEnforcement(t *testing.T) {}
func TestQuotaEnforcement(t *testing.T) {}

// Python
def test_rag_retrieval():
def test_tool_execution():
def test_embedding_generation():
```

### Testes de Carga

```go
// 1000 agentes concorrentes
// 10000 mensagens/minuto
// Validação de rate limit
// Monitoramento de saúde dos workers Python
```

---

# 📊 OBSERVABILIDADE

## Métricas (Prometheus)

| Métrica | Tipo | Labels | Camada |
|---------|------|--------|--------|
| `aiox_agents_created_total` | Counter | user_id | Go |
| `aiox_memory_bytes` | Gauge | agent_id, user_id | Go |
| `aiox_tools_executed_total` | Counter | tool_name, agent_id | Python |
| `aiox_tokens_used_total` | Counter | agent_id, user_id | Python |
| `aiox_latency_seconds` | Histogram | agent_id, operation | Ambos |
| `aiox_policy_violations_total` | Counter | violation_type, severity | Go |
| `aiox_quota_remaining` | Gauge | user_id | Go |
| `aiox_python_worker_health` | Gauge | worker_id, status | Go |
| `aiox_grpc_request_duration` | Histogram | method, status | Ambos |
| `aiox_llm_inference_time` | Histogram | model, provider | Python |

## Logs (Structured)

```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "level": "info",
  "event": "agent_message_processed",
  "owner_user_id": "uuid",
  "agent_id": "uuid",
  "go_latency_ms": 15,
  "python_latency_ms": 135,
  "worker_id": "python-worker-1",
  "tokens_used": 256,
  "policy_checked": true
}
```

## Tracing (OpenTelemetry)

```
Trace completo por mensagem:
├── [Go] API Gateway Receive
├── [Go] Auth Validation
├── [Go] Ownership Check
├── [Go] Policy Check
├── [Go] Queue Prioritization
├── [Go→Python] gRPC Dispatch
├── [Python] LLM Inference
├── [Python] RAG Retrieval
├── [Python] Tool Execution
├── [Python→Go] gRPC Response
├── [Go] Memory Update (Redis)
├── [Go] Audit Log
└── [Go] XMPP Response
```

---

# 🎯 RESULTADO FINAL

Uma plataforma:

* ✅ Multi-agente federável
* ✅ Com identidade XMPP real
* ✅ Cluster-ready
* ✅ Determinística
* ✅ Escalável
* ✅ Com UI de criação
* ✅ Multi-tenant isolada
* ✅ Com memória vetorial segura
* ✅ Com governance e policy engine
* ✅ Com audit log completo
* ✅ **Arquitetura híbrida Go + Python otimizada**
* ✅ **Comunicação gRPC de baixa latência**
* ✅ **Workers Python auto-scaláveis**
* ✅ Pronta para produção

---

# 📋 FASES DE IMPLEMENTAÇÃO

## Fase 1 — Fundação Go (Semana 1-2)
- [ ] DB Schema completo
- [ ] Auth Service (JWT + XMPP SASL)
- [ ] CRUD de Agentes (API Go)
- [ ] Validação de Ownership básica
- [ ] Estrutura de pastas Go

## Fase 2 — XMPP + Orquestração Go (Semana 3-4)
- [ ] XMPP Component integration (Go)
- [ ] Message routing
- [ ] Ownership validation no fluxo XMPP
- [ ] Orchestrator básico (Go)
- [ ] NATS event bus

## Fase 3 — AI Worker Python (Semana 5-6)
- [ ] **Setup do worker Python**
- [ ] **Integração gRPC Go ↔ Python**
- [ ] **LLM integration (LangChain/OpenAI)**
- [ ] **RAG com pgvector (Python)**
- [ ] **Embedding generation**

## Fase 4 — Memória + Tools (Semana 7-8)
- [ ] Redis short-term memory (Go)
- [ ] Postgres + pgvector long-term (Python)
- [ ] Tool registry (Go definitions)
- [ ] Tool implementations (Python)
- [ ] Vector query isolation

## Fase 5 — Governance + UI (Semana 9-10)
- [ ] Policy engine (Go)
- [ ] Quota enforcement (Go)
- [ ] Audit logging (Go)
- [ ] Web UI completa
- [ ] Dashboard de governança
- [ ] Worker health monitoring

## Fase 6 — Produção (Semana 11-12)
- [ ] Docker compose completo
- [ ] Testes de segurança
- [ ] Testes de carga
- [ ] Observabilidade
- [ ] Auto-scaling de workers Python
- [ ] Documentação

---

# 🚀 PRÓXIMOS PASSOS (ROADMAP)

Após esta versão estabilizada:

| Prioridade | Feature | Valor |
|------------|---------|-------|
| 1 | Federation XMPP entre hubs | Escala horizontal |
| 2 | Marketplace público de agentes | Monetização |
| 3 | Token billing + payments | Revenue |
| 4 | Versionamento de agentes (v1, v2, rollback) | DevEx |
| 5 | Agent-to-Agent communication | Autonomia |
| 6 | Multi-model LLM routing | Cost optimization |

---

# ⚠️ INSTRUÇÕES PARA CLAUDE CODE

1. **Comece pela Fase 1** (DB + Auth + CRUD em Go)
2. **Valide isolamento** antes de prosseguir
3. **Gere testes de segurança** junto com o código
4. **Use contextos separados** para Go e Python
5. **Mantenha compatibilidade** com RFC AIOX original
6. **Priorize comunicação gRPC** entre Go e Python
7. **Garanta que Python nunca receba dados sem owner_user_id**
8. **Guarde o prompt inicial**
9. **Para cada passo escreva o que foi feito e o que falta para o proximos passos**

---

**Agora você está construindo:**

> "Infraestrutura federável de agentes privados configuráveis pelo usuário, com arquitetura híbrida Go (plataforma) + Python (IA), governance enterprise-ready"

