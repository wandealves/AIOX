package xmpp

import (
	"context"
	"log/slog"

	"gosrc.io/xmpp"

	"github.com/aiox-platform/aiox/internal/config"
)

// Component manages the XMPP external component lifecycle (XEP-0114).
type Component struct {
	sm     *xmpp.StreamManager
	comp   *xmpp.Component
	cancel context.CancelFunc
}

// NewComponent creates a new XMPP component with the given handler.
func NewComponent(cfg config.XMPPConfig, handler *Handler) (*Component, error) {
	router := xmpp.NewRouter()
	router.HandleFunc("message", handler.HandleMessage)
	router.HandleFunc("presence", handler.HandlePresence)
	router.HandleFunc("iq", handler.HandleIQ)

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
		slog.Error("XMPP component error", "error", err)
	})
	if err != nil {
		return nil, err
	}

	sm := xmpp.NewStreamManager(comp, func(s xmpp.Sender) {
		slog.Info("XMPP component connected", "domain", cfg.ComponentName)
	})

	return &Component{sm: sm, comp: comp}, nil
}

// Start runs the XMPP component. It blocks until the context is cancelled or an error occurs.
func (c *Component) Start(ctx context.Context) error {
	ctx, c.cancel = context.WithCancel(ctx)

	errCh := make(chan error, 1)
	go func() {
		errCh <- c.sm.Run()
	}()

	select {
	case <-ctx.Done():
		c.sm.Stop()
		return nil
	case err := <-errCh:
		return err
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
	c.sm.Stop()
}
