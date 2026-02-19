package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/aiox-platform/aiox/internal/agents"
	"github.com/aiox-platform/aiox/internal/governance"
	"github.com/aiox-platform/aiox/internal/governance/quota"
	"github.com/aiox-platform/aiox/internal/memory"
	inats "github.com/aiox-platform/aiox/internal/nats"
	pb "github.com/aiox-platform/aiox/internal/worker/workerpb"
)

// pendingTask holds metadata for a dispatched task awaiting a response.
type pendingTask struct {
	RequestID    string
	AgentID      uuid.UUID
	OwnerUserID  uuid.UUID
	FromJID      string
	AgentJID     string
	AgentName    string
	WorkerID     string
	Input        string
	DispatchedAt time.Time
	MemoryConfig memory.MemoryConfig
}

// Dispatcher consumes tasks from NATS, dispatches to Python workers via gRPC,
// and publishes outbound messages when workers return results.
type Dispatcher struct {
	pool        *Pool
	publisher   *inats.Publisher
	consumerMgr *inats.ConsumerManager
	agentSvc    *agents.Service
	repo        *Repository
	memorySvc   *memory.Service
	quotaSvc    *quota.Service
	resultCh    <-chan *pb.TaskResponse
	taskTimeout time.Duration

	mu      sync.Mutex
	pending map[string]*pendingTask
}

// NewDispatcher creates a new task dispatcher.
func NewDispatcher(
	pool *Pool,
	publisher *inats.Publisher,
	consumerMgr *inats.ConsumerManager,
	agentSvc *agents.Service,
	repo *Repository,
	memorySvc *memory.Service,
	quotaSvc *quota.Service,
	resultCh <-chan *pb.TaskResponse,
	taskTimeoutSec int,
) *Dispatcher {
	timeout := time.Duration(taskTimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &Dispatcher{
		pool:        pool,
		publisher:   publisher,
		consumerMgr: consumerMgr,
		agentSvc:    agentSvc,
		repo:        repo,
		memorySvc:   memorySvc,
		quotaSvc:    quotaSvc,
		resultCh:    resultCh,
		taskTimeout: timeout,
		pending:     make(map[string]*pendingTask),
	}
}

// Start runs the dispatcher's consume, result processing, and timeout cleanup loops.
func (d *Dispatcher) Start(ctx context.Context) error {
	consumer, err := d.consumerMgr.EnsureConsumer(ctx, inats.StreamTasks, "task-dispatcher", "aiox.tasks.>")
	if err != nil {
		return err
	}

	slog.Info("task dispatcher started", "timeout", d.taskTimeout)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		d.consumeTasks(ctx, consumer)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		d.processResults(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		d.cleanupTimeouts(ctx)
	}()

	wg.Wait()
	return nil
}

func (d *Dispatcher) consumeTasks(ctx context.Context, consumer jetstream.Consumer) {
	for {
		msgs, err := consumer.Fetch(10, jetstream.FetchMaxWait(inats.FetchTimeout))
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Debug("dispatcher: fetching tasks", "error", err)
			continue
		}

		for msg := range msgs.Messages() {
			d.handleTask(ctx, msg)
		}

		if ctx.Err() != nil {
			return
		}
	}
}

func (d *Dispatcher) handleTask(ctx context.Context, msg jetstream.Msg) {
	var task inats.TaskMessage
	if err := json.Unmarshal(msg.Data(), &task); err != nil {
		slog.Error("dispatcher: unmarshaling task", "error", err)
		_ = msg.Nak()
		return
	}

	// Fetch agent to get decrypted system prompt and LLM config
	agent, err := d.agentSvc.GetByID(ctx, task.AgentID)
	if err != nil {
		slog.Error("dispatcher: fetching agent", "error", err, "agent_id", task.AgentID)
		_ = msg.Nak()
		return
	}
	if agent == nil {
		slog.Warn("dispatcher: agent not found", "agent_id", task.AgentID)
		d.sendErrorResponse(ctx, task, "Agent not found")
		_ = msg.Ack()
		return
	}

	// Governance checks at dispatch time
	gov := governance.ParseGovernance(agent.Governance)

	if gov.Blocked {
		slog.Warn("dispatcher: agent blocked by governance", "agent_id", task.AgentID)
		d.sendErrorResponse(ctx, task, "Agent is blocked by governance policy")
		_ = msg.Ack()
		return
	}

	// Check allowed providers against agent's LLM config
	if len(gov.AllowedProviders) > 0 {
		provider := extractProvider(agent.LLMConfig)
		if provider != "" && !providerAllowed(provider, gov.AllowedProviders) {
			slog.Warn("dispatcher: provider not allowed", "agent_id", task.AgentID, "provider", provider)
			d.sendErrorResponse(ctx, task, "LLM provider '"+provider+"' not allowed by governance policy")
			_ = msg.Ack()
			return
		}
	}

	// Select a worker
	worker := d.pool.SelectWorker()
	if worker == nil {
		slog.Warn("dispatcher: no workers available, nacking for retry", "request_id", task.RequestID)
		_ = msg.Nak()
		return
	}

	// Build task request
	llmConfigJSON, _ := json.Marshal(json.RawMessage(agent.LLMConfig))

	taskReq := &pb.TaskRequest{
		RequestId:     task.RequestID,
		AgentId:       task.AgentID.String(),
		OwnerUserId:   task.OwnerUserID.String(),
		UserMessage:   task.Message,
		SystemPrompt:  agent.Profile.SystemPrompt,
		LlmConfigJson: string(llmConfigJSON),
		FromJid:       task.FromJID,
		AgentJid:      task.AgentJID,
		AgentName:     task.AgentName,
	}

	// Parse memory config and fetch conversation context
	memCfg := memory.ParseConfig(agent.MemoryConfig)
	if memCfg.Enabled && d.memorySvc != nil {
		// Note: queryEmbedding is nil here — on the first message there are no prior
		// embeddings, so long-term search returns empty. Embeddings are generated by
		// the Python worker and stored after the response. On subsequent messages the
		// dispatcher still passes nil because embedding generation only happens in Python.
		// Future: could cache the last user embedding in Redis for retrieval here.
		memCtx, err := d.memorySvc.GetConversationContext(
			ctx, task.AgentID, task.OwnerUserID, task.FromJID, memCfg, nil,
		)
		if err != nil {
			slog.Warn("dispatcher: fetching memory context", "error", err, "agent_id", task.AgentID)
		} else if memCtx != nil {
			if ctxJSON, err := json.Marshal(memCtx); err == nil {
				taskReq.MemoryContextJson = string(ctxJSON)
			}
		}

		if cfgJSON, err := json.Marshal(memCfg); err == nil {
			taskReq.MemoryConfigJson = string(cfgJSON)
		}
	}

	// Send to worker
	if err := worker.Send(&pb.ServerMessage{
		Payload: &pb.ServerMessage_TaskRequest{
			TaskRequest: taskReq,
		},
	}); err != nil {
		slog.Error("dispatcher: sending task to worker", "error", err, "worker_id", worker.WorkerID)
		_ = msg.Nak()
		return
	}

	worker.IncrementActive()

	// Track pending task
	d.mu.Lock()
	d.pending[task.RequestID] = &pendingTask{
		RequestID:    task.RequestID,
		AgentID:      task.AgentID,
		OwnerUserID:  task.OwnerUserID,
		FromJID:      task.FromJID,
		AgentJID:     task.AgentJID,
		AgentName:    task.AgentName,
		WorkerID:     worker.WorkerID,
		Input:        task.Message,
		DispatchedAt: time.Now(),
		MemoryConfig: memCfg,
	}
	d.mu.Unlock()

	_ = msg.Ack()

	slog.Debug("dispatcher: task dispatched",
		"request_id", task.RequestID,
		"agent_id", task.AgentID,
		"worker_id", worker.WorkerID,
	)
}

func (d *Dispatcher) processResults(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case resp := <-d.resultCh:
			d.handleResult(ctx, resp)
		}
	}
}

func (d *Dispatcher) handleResult(ctx context.Context, resp *pb.TaskResponse) {
	d.mu.Lock()
	pt, ok := d.pending[resp.RequestId]
	if ok {
		delete(d.pending, resp.RequestId)
	}
	d.mu.Unlock()

	if !ok {
		slog.Warn("dispatcher: received result for unknown request", "request_id", resp.RequestId)
		return
	}

	// Decrement worker's active count
	if w := d.pool.Get(resp.WorkerId); w != nil {
		w.DecrementActive()
	}

	goLatency := int(time.Since(pt.DispatchedAt).Milliseconds())

	// Determine response body
	body := resp.ResponseText
	status := "completed"
	if resp.ErrorMessage != "" {
		body = "Error processing your message: " + resp.ErrorMessage
		status = "error"
	}

	// Publish outbound message
	outbound := inats.OutboundMessage{
		ID:        uuid.New().String(),
		ToJID:     pt.FromJID,
		FromJID:   pt.AgentJID,
		Body:      body,
		InReplyTo: pt.RequestID,
	}
	if err := d.publisher.PublishOutboundMessage(ctx, outbound); err != nil {
		slog.Error("dispatcher: publishing outbound", "error", err)
	}

	// Record execution
	exec := &Execution{
		ID:              uuid.New(),
		OwnerUserID:     pt.OwnerUserID,
		AgentID:         pt.AgentID,
		Input:           pt.Input,
		Output:          resp.ResponseText,
		TokensUsed:      int(resp.TokensUsed),
		WorkerID:        resp.WorkerId,
		DurationMs:      int(resp.DurationMs),
		GoLatencyMs:     goLatency,
		PythonLatencyMs: int(resp.DurationMs),
		Status:          status,
		ErrorMessage:    resp.ErrorMessage,
		CreatedAt:       time.Now(),
	}
	if err := d.repo.RecordExecution(ctx, exec); err != nil {
		slog.Error("dispatcher: recording execution", "error", err)
	}

	// Deduct tokens from quota after successful completion
	if status == "completed" && resp.TokensUsed > 0 && d.quotaSvc != nil {
		if err := d.quotaSvc.DeductTokens(ctx, pt.OwnerUserID, int(resp.TokensUsed)); err != nil {
			slog.Warn("dispatcher: deducting tokens from quota", "error", err, "user_id", pt.OwnerUserID)
		}
	}

	// Store memory if enabled
	if pt.MemoryConfig.Enabled && d.memorySvc != nil && status == "completed" {
		// Store short-term conversation turn
		if err := d.memorySvc.StoreConversationTurn(ctx, pt.AgentID, pt.FromJID, pt.Input, resp.ResponseText, pt.MemoryConfig); err != nil {
			slog.Warn("dispatcher: storing conversation turn", "error", err, "agent_id", pt.AgentID)
		}

		// Store long-term memories returned by the Python worker (with embeddings)
		if pt.MemoryConfig.LongTermEnabled {
			for _, mem := range resp.NewMemories {
				embedding := make([]float32, len(mem.Embedding))
				copy(embedding, mem.Embedding)

				metadata := json.RawMessage(`{}`)
				if mem.MetadataJson != "" {
					metadata = json.RawMessage(mem.MetadataJson)
				}

				m := &memory.Memory{
					OwnerUserID: pt.OwnerUserID,
					AgentID:     pt.AgentID,
					Content:     mem.Content,
					Embedding:   embedding,
					MemoryType:  mem.MemoryType,
					Metadata:    metadata,
				}
				if err := d.memorySvc.StoreLongTermMemory(ctx, m); err != nil {
					slog.Warn("dispatcher: storing long-term memory", "error", err, "agent_id", pt.AgentID)
				}
			}
		}
	}

	// Audit event
	audit := inats.AuditEvent{
		OwnerUserID:  pt.OwnerUserID,
		EventType:    "task_completed",
		Severity:     "info",
		ResourceType: "agent",
		ResourceID:   pt.AgentID.String(),
		Details:      "Task processed by worker " + resp.WorkerId + ", model: " + resp.ModelUsed,
		Timestamp:    time.Now().UTC(),
	}
	if status == "error" {
		audit.Severity = "warn"
		audit.EventType = "task_failed"
	}
	if err := d.publisher.PublishAuditEvent(ctx, audit); err != nil {
		slog.Error("dispatcher: publishing audit event", "error", err)
	}

	slog.Debug("dispatcher: result processed",
		"request_id", resp.RequestId,
		"worker_id", resp.WorkerId,
		"status", status,
		"tokens", resp.TokensUsed,
		"duration_ms", resp.DurationMs,
	)
}

func (d *Dispatcher) cleanupTimeouts(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.expireStale(ctx)
		}
	}
}

func (d *Dispatcher) expireStale(ctx context.Context) {
	d.mu.Lock()
	var expired []*pendingTask
	now := time.Now()
	for id, pt := range d.pending {
		if now.Sub(pt.DispatchedAt) > d.taskTimeout {
			expired = append(expired, pt)
			delete(d.pending, id)
		}
	}
	d.mu.Unlock()

	for _, pt := range expired {
		slog.Warn("dispatcher: task timed out", "request_id", pt.RequestID, "agent_id", pt.AgentID)

		// Send timeout error to user
		outbound := inats.OutboundMessage{
			ID:        uuid.New().String(),
			ToJID:     pt.FromJID,
			FromJID:   pt.AgentJID,
			Body:      "Sorry, the request timed out. Please try again.",
			InReplyTo: pt.RequestID,
		}
		if err := d.publisher.PublishOutboundMessage(ctx, outbound); err != nil {
			slog.Error("dispatcher: publishing timeout response", "error", err)
		}

		// Record failed execution
		exec := &Execution{
			ID:           uuid.New(),
			OwnerUserID:  pt.OwnerUserID,
			AgentID:      pt.AgentID,
			Input:        pt.Input,
			Status:       "timeout",
			ErrorMessage: "task timed out after " + d.taskTimeout.String(),
			WorkerID:     pt.WorkerID,
			GoLatencyMs:  int(time.Since(pt.DispatchedAt).Milliseconds()),
			CreatedAt:    time.Now(),
		}
		if err := d.repo.RecordExecution(ctx, exec); err != nil {
			slog.Error("dispatcher: recording timeout execution", "error", err)
		}

		// Decrement worker active count
		if w := d.pool.Get(pt.WorkerID); w != nil {
			w.DecrementActive()
		}
	}
}

func (d *Dispatcher) sendErrorResponse(ctx context.Context, task inats.TaskMessage, errMsg string) {
	outbound := inats.OutboundMessage{
		ID:        uuid.New().String(),
		ToJID:     task.FromJID,
		FromJID:   task.AgentJID,
		Body:      "Error: " + errMsg,
		InReplyTo: task.RequestID,
	}
	if err := d.publisher.PublishOutboundMessage(ctx, outbound); err != nil {
		slog.Error("dispatcher: publishing error response", "error", err)
	}
}

// extractProvider parses the provider field from the LLM config JSON.
func extractProvider(llmConfig json.RawMessage) string {
	if len(llmConfig) == 0 {
		return ""
	}
	var cfg struct {
		Provider string `json:"provider"`
	}
	if err := json.Unmarshal(llmConfig, &cfg); err != nil {
		return ""
	}
	return cfg.Provider
}

// providerAllowed checks if a provider is in the allowed list (case-insensitive).
func providerAllowed(provider string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(a, provider) {
			return true
		}
	}
	return false
}
