//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/aiox-platform/aiox/internal/agents"
	"github.com/aiox-platform/aiox/internal/config"
	inats "github.com/aiox-platform/aiox/internal/nats"
	"github.com/aiox-platform/aiox/internal/orchestrator"
)

func TestOrchestratorFlow(t *testing.T) {
	// Setup test env (Postgres + Redis)
	env := SetupTestEnv(t)
	ctx := context.Background()

	// Setup NATS container
	natsContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "nats:2-alpine",
			ExposedPorts: []string{"4222/tcp"},
			Cmd:          []string{"--jetstream", "--store_dir", "/data"},
			WaitingFor:   wait.ForLog("Server is ready").WithStartupTimeout(30 * time.Second),
		},
		Started: true,
	})
	require.NoError(t, err)
	t.Cleanup(func() { natsContainer.Terminate(ctx) })

	host, _ := natsContainer.Host(ctx)
	port, _ := natsContainer.MappedPort(ctx, "4222")

	natsClient, err := inats.NewClient(ctx, config.NATSConfig{
		URL: fmt.Sprintf("nats://%s:%s", host, port.Port()),
	})
	require.NoError(t, err)
	t.Cleanup(func() { natsClient.Close() })

	publisher := inats.NewPublisher(natsClient.JetStream())
	consumerMgr := inats.NewConsumerManager(natsClient.JetStream())

	// Create a user and agent via HTTP API
	RegisterUser(t, env, "orch-test@aiox.local", "password123")
	token := LoginUser(t, env, "orch-test@aiox.local", "password123")

	agentBody := map[string]any{
		"name":          "Orchestrator Test Agent",
		"system_prompt": "You are a test agent",
	}
	resp := DoRequest(t, env, "POST", "/api/v1/agents", agentBody, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	agentResp := ParseResponse(t, resp)
	agentData := agentResp["data"].(map[string]any)
	agentJID := agentData["jid"].(string)

	// Setup orchestrator
	agentRepo := agents.NewRepository(env.Pool)
	validator := orchestrator.NewValidator()
	orchRouter := orchestrator.NewRouter(agentRepo)
	orch := orchestrator.NewOrchestrator(publisher, consumerMgr, validator, orchRouter)

	// Start orchestrator in background
	orchCtx, orchCancel := context.WithCancel(ctx)
	defer orchCancel()

	orchDone := make(chan error, 1)
	go func() {
		orchDone <- orch.Start(orchCtx)
	}()

	// Wait for orchestrator to start
	time.Sleep(500 * time.Millisecond)

	// Publish an inbound message as if it came from XMPP
	inbound := inats.InboundMessage{
		ID:         "orch-test-1",
		FromJID:    "user@aiox.local/resource",
		ToJID:      agentJID,
		Body:       "Hello agent!",
		StanzaType: "chat",
		ReceivedAt: time.Now().UTC(),
	}
	err = publisher.PublishInboundMessage(ctx, inbound)
	require.NoError(t, err)

	// Consume the outbound response
	outConsumer, err := consumerMgr.EnsureConsumer(ctx, inats.StreamMessages, "test-outbound", inats.SubjectOutboundMessage)
	require.NoError(t, err)

	var outbound inats.OutboundMessage
	deadline := time.After(10 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timed out waiting for outbound message")
		default:
		}

		msgs, err := outConsumer.Fetch(1, jetstream.FetchMaxWait(2*time.Second))
		if err != nil {
			continue
		}
		for m := range msgs.Messages() {
			err = json.Unmarshal(m.Data(), &outbound)
			require.NoError(t, err)
			_ = m.Ack()
		}
		if outbound.ID != "" {
			break
		}
	}

	// Verify the outbound response
	assert.Equal(t, "user@aiox.local/resource", outbound.ToJID)
	assert.Equal(t, agentJID, outbound.FromJID)
	assert.Contains(t, outbound.Body, "Orchestrator Test Agent")
	assert.Equal(t, "orch-test-1", outbound.InReplyTo)

	// Cleanup
	orchCancel()
	select {
	case err := <-orchDone:
		assert.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("orchestrator did not stop in time")
	}
}
