package nats

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"

	"github.com/aiox-platform/aiox/internal/config"
)

// Client wraps a NATS connection with JetStream support.
type Client struct {
	conn *nats.Conn
	js   jetstream.JetStream
}

// NewClient connects to NATS and ensures required JetStream streams exist.
func NewClient(ctx context.Context, cfg config.NATSConfig) (*Client, error) {
	nc, err := nats.Connect(cfg.URL,
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			slog.Warn("NATS disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			slog.Info("NATS reconnected")
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("connecting to NATS: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("creating JetStream context: %w", err)
	}

	c := &Client{conn: nc, js: js}

	if err := c.ensureStreams(ctx); err != nil {
		nc.Close()
		return nil, fmt.Errorf("ensuring streams: %w", err)
	}

	slog.Info("connected to NATS", "url", cfg.URL)
	return c, nil
}

func (c *Client) ensureStreams(ctx context.Context) error {
	streams := []jetstream.StreamConfig{
		{
			Name:      StreamMessages,
			Subjects:  []string{"aiox.messages.>"},
			Retention: jetstream.WorkQueuePolicy,
			MaxAge:    24 * time.Hour,
		},
		{
			Name:      StreamTasks,
			Subjects:  []string{"aiox.tasks.>"},
			Retention: jetstream.WorkQueuePolicy,
			MaxAge:    1 * time.Hour,
		},
		{
			Name:      StreamEvents,
			Subjects:  []string{"aiox.events.>"},
			Retention: jetstream.LimitsPolicy,
			MaxAge:    7 * 24 * time.Hour,
		},
	}

	for _, cfg := range streams {
		_, err := c.js.CreateOrUpdateStream(ctx, cfg)
		if err != nil {
			return fmt.Errorf("creating stream %s: %w", cfg.Name, err)
		}
		slog.Debug("ensured NATS stream", "name", cfg.Name)
	}
	return nil
}

// JetStream returns the JetStream context.
func (c *Client) JetStream() jetstream.JetStream {
	return c.js
}

// Conn returns the underlying NATS connection.
func (c *Client) Conn() *nats.Conn {
	return c.conn
}

// Healthy returns true if NATS connection is active.
func (c *Client) Healthy() bool {
	return c.conn.IsConnected()
}

// Close drains and closes the NATS connection.
func (c *Client) Close() {
	if err := c.conn.Drain(); err != nil {
		slog.Warn("draining NATS connection", "error", err)
	}
}
