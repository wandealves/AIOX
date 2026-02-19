//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/aiox-platform/aiox/internal/config"
	inats "github.com/aiox-platform/aiox/internal/nats"
)

func setupNATSContainer(t *testing.T) *inats.Client {
	t.Helper()
	ctx := context.Background()

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

	client, err := inats.NewClient(ctx, config.NATSConfig{
		URL: fmt.Sprintf("nats://%s:%s", host, port.Port()),
	})
	require.NoError(t, err)
	t.Cleanup(func() { client.Close() })

	return client
}

func TestNATSPublishConsume(t *testing.T) {
	client := setupNATSContainer(t)
	ctx := context.Background()

	publisher := inats.NewPublisher(client.JetStream())
	consumerMgr := inats.NewConsumerManager(client.JetStream())

	t.Run("publish and consume inbound message", func(t *testing.T) {
		msg := inats.InboundMessage{
			ID:         "test-msg-1",
			FromJID:    "user@aiox.local",
			ToJID:      "agent-123@agents.aiox.local",
			Body:       "hello agent",
			StanzaType: "chat",
			ReceivedAt: time.Now().UTC(),
		}

		err := publisher.PublishInboundMessage(ctx, msg)
		require.NoError(t, err)

		consumer, err := consumerMgr.EnsureConsumer(ctx, inats.StreamMessages, "test-consumer", inats.SubjectInboundMessage)
		require.NoError(t, err)

		msgs, err := consumer.Fetch(1, jetstream.FetchMaxWait(5*time.Second))
		require.NoError(t, err)

		var received inats.InboundMessage
		for m := range msgs.Messages() {
			err = json.Unmarshal(m.Data(), &received)
			require.NoError(t, err)
			_ = m.Ack()
		}

		assert.Equal(t, "test-msg-1", received.ID)
		assert.Equal(t, "hello agent", received.Body)
		assert.Equal(t, "user@aiox.local", received.FromJID)
	})

	t.Run("NATS client is healthy", func(t *testing.T) {
		assert.True(t, client.Healthy())
	})
}
