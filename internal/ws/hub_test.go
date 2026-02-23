package ws

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHub_RegisterAndSend(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	conn := &Conn{
		ID:     "conn-1",
		UserID: "user-1",
		hub:    hub,
		send:   make(chan []byte, sendBufLen),
		done:   make(chan struct{}),
	}

	hub.Register(conn)
	time.Sleep(50 * time.Millisecond) // let hub process

	assert.Equal(t, 1, hub.ConnectedCount())

	hub.SendToUser("user-1", []byte(`{"test":"data"}`))

	select {
	case msg := <-conn.send:
		assert.Equal(t, `{"test":"data"}`, string(msg))
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestHub_UnregisterClosesChannel(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	conn := &Conn{
		ID:     "conn-2",
		UserID: "user-2",
		hub:    hub,
		send:   make(chan []byte, sendBufLen),
		done:   make(chan struct{}),
	}

	hub.Register(conn)
	time.Sleep(50 * time.Millisecond)

	hub.Unregister(conn)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 0, hub.ConnectedCount())

	// done channel should be closed
	select {
	case <-conn.done:
		// expected
	default:
		t.Fatal("done channel should be closed after unregister")
	}
}

func TestHub_MultipleConnectionsPerUser(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	conn1 := &Conn{
		ID:     "conn-1",
		UserID: "user-1",
		hub:    hub,
		send:   make(chan []byte, sendBufLen),
		done:   make(chan struct{}),
	}
	conn2 := &Conn{
		ID:     "conn-2",
		UserID: "user-1",
		hub:    hub,
		send:   make(chan []byte, sendBufLen),
		done:   make(chan struct{}),
	}

	hub.Register(conn1)
	hub.Register(conn2)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, 2, hub.ConnectedCount())

	hub.SendToUser("user-1", []byte(`hello`))

	// Both connections should receive the message
	for _, c := range []*Conn{conn1, conn2} {
		select {
		case msg := <-c.send:
			assert.Equal(t, "hello", string(msg))
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message on connection", c.ID)
		}
	}
}

func TestHub_SendToNonexistentUser(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	// Should not panic
	hub.SendToUser("nonexistent", []byte(`test`))
}

func TestHub_ConcurrentOperations(t *testing.T) {
	hub := NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go hub.Run(ctx)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			userID := "user-1"
			c := &Conn{
				ID:     "conn-" + string(rune('A'+n)),
				UserID: userID,
				hub:    hub,
				send:   make(chan []byte, sendBufLen),
				done:   make(chan struct{}),
			}
			hub.Register(c)
			time.Sleep(10 * time.Millisecond)
			hub.SendToUser(userID, []byte(`msg`))
			time.Sleep(10 * time.Millisecond)
			hub.Unregister(c)
		}(i)
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)
}

func TestExtractUserIDFromJID(t *testing.T) {
	tests := []struct {
		jid      string
		expected string
	}{
		{"ws-abc123@ws.local", "abc123"},
		{"ws-abc123@ws.local/resource", "abc123"},
		{"agent-abc@agents.local", ""},
		{"nope", ""},
		{"ws-@ws.local", ""},
	}

	for _, tt := range tests {
		result := extractUserIDFromJID(tt.jid)
		assert.Equal(t, tt.expected, result, "jid=%s", tt.jid)
	}
}

func TestExtractAgentIDFromJID(t *testing.T) {
	tests := []struct {
		jid      string
		expected string
	}{
		{"agent-abc-123@agents.local", "abc-123"},
		{"agent-abc-123@agents.local/resource", "abc-123"},
		{"user@domain", "user"},
		{"nope", ""},
	}

	for _, tt := range tests {
		result := extractAgentIDFromJID(tt.jid)
		require.Equal(t, tt.expected, result, "jid=%s", tt.jid)
	}
}
