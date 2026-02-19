package xmpp

import (
	"context"
	"log/slog"
	"time"

	"gosrc.io/xmpp"

	"github.com/aiox-platform/aiox/internal/config"
)

const reconnectDelay = 5 * time.Second

// Component manages the XMPP external component lifecycle (XEP-0114).
//
// NOTE: gosrc.io/xmpp StreamManager does a type assertion to *xmpp.Client
// internally, so it never works for *xmpp.Component. We manage the connect /
// reconnect loop ourselves instead.
type Component struct {
	comp        *xmpp.Component
	reconnectCh chan struct{}
	cancel      context.CancelFunc
}

// NewComponent creates a new XMPP component with the given handler.
func NewComponent(cfg config.XMPPConfig, handler *Handler) (*Component, error) {
	router := xmpp.NewRouter()
	router.HandleFunc("message", handler.HandleMessage)
	router.HandleFunc("presence", handler.HandlePresence)
	router.HandleFunc("iq", handler.HandleIQ)

	reconnectCh := make(chan struct{}, 1)

	opts := xmpp.ComponentOptions{
		TransportConfiguration: xmpp.TransportConfiguration{
			Address: cfg.ComponentAddr(),
			Domain:  cfg.ComponentName,
		},
		Domain:   cfg.ComponentName,
		Secret:   cfg.ComponentSecret,
		Name:     "AIOX Agent Gateway",
		Category: "gateway",
		Type:     "service",
	}

	comp, err := xmpp.NewComponent(opts, router, func(err error) {
		slog.Error("XMPP component stream error", "error", err)
		// Signal the Start loop to reconnect.
		select {
		case reconnectCh <- struct{}{}:
		default:
		}
	})
	if err != nil {
		return nil, err
	}

	return &Component{comp: comp, reconnectCh: reconnectCh}, nil
}

// Start runs the XMPP component with automatic reconnection.
// It blocks until ctx is cancelled.
func (c *Component) Start(ctx context.Context) error {
	ctx, c.cancel = context.WithCancel(ctx)

	for {
		slog.Info("XMPP component connecting")
		if err := c.comp.Connect(); err != nil {
			slog.Error("XMPP component connect failed", "error", err)
		} else {
			slog.Info("XMPP component connected")
		}

		// Wait for a disconnection event or shutdown signal.
		select {
		case <-ctx.Done():
			_ = c.comp.Disconnect()
			return nil
		case <-c.reconnectCh:
			slog.Info("XMPP component reconnecting", "delay", reconnectDelay)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(reconnectDelay):
				// loop → reconnect
			}
		}
	}
}

// Sender returns the underlying component for sending stanzas.
func (c *Component) Sender() xmpp.Sender {
	return c.comp
}

// Stop disconnects the XMPP component.
func (c *Component) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	_ = c.comp.Disconnect()
}
