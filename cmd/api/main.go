package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/aiox-platform/aiox/internal/agents"
	"github.com/aiox-platform/aiox/internal/api"
	"github.com/aiox-platform/aiox/internal/auth"
	"github.com/aiox-platform/aiox/internal/config"
	"github.com/aiox-platform/aiox/internal/database"
	"github.com/aiox-platform/aiox/internal/governance"
	"github.com/aiox-platform/aiox/internal/governance/audit"
	"github.com/aiox-platform/aiox/internal/governance/quota"
	"github.com/aiox-platform/aiox/internal/memory"
	"github.com/aiox-platform/aiox/internal/middleware"
	inats "github.com/aiox-platform/aiox/internal/nats"
	"github.com/aiox-platform/aiox/internal/orchestrator"
	iredis "github.com/aiox-platform/aiox/internal/redis"
	"github.com/aiox-platform/aiox/internal/server"
	"github.com/aiox-platform/aiox/internal/users"
	"github.com/aiox-platform/aiox/internal/worker"
	pb "github.com/aiox-platform/aiox/internal/worker/workerpb"
	ixmpp "github.com/aiox-platform/aiox/internal/xmpp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("loading config", "error", err)
		os.Exit(1)
	}

	setupLogger(cfg.Log)

	if err := cfg.Validate(); err != nil {
		slog.Error("config validation failed", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Auto-migrate if enabled
	if cfg.DB.AutoMigrate {
		slog.Info("running database migrations", "path", cfg.DB.MigrationsPath)
		if err := database.RunMigrations(cfg.DB.DSN(), cfg.DB.MigrationsPath); err != nil {
			slog.Error("auto-migration failed", "error", err)
			os.Exit(1)
		}
	}

	// PostgreSQL
	pool, err := database.NewPostgresPool(ctx, cfg.DB)
	if err != nil {
		slog.Error("connecting to postgres", "error", err)
		os.Exit(1)
	}

	// Redis
	redisClient, err := iredis.NewClient(ctx, cfg.Redis)
	if err != nil {
		slog.Error("connecting to redis", "error", err)
		os.Exit(1)
	}

	// NATS
	natsClient, err := inats.NewClient(ctx, cfg.NATS)
	if err != nil {
		slog.Error("connecting to nats", "error", err)
		os.Exit(1)
	}

	// Auth
	jwtManager := auth.NewJWTManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessExpiry,
		cfg.JWT.RefreshExpiry,
	)
	authSvc := auth.NewService(jwtManager, redisClient)
	userRepo := users.NewRepository(pool)
	userSvc := users.NewService(userRepo)
	authHandler := auth.NewHandler(authSvc, userSvc)

	// Agents
	agentRepo := agents.NewRepository(pool)
	agentSvc := agents.NewService(agentRepo, cfg.Encryption.Key, cfg.XMPP.Domain)
	agentHandler := agents.NewHandler(agentSvc)

	// Memory (Phase 4)
	memoryRepo := memory.NewPostgresRepository(pool)
	shortTermStore := memory.NewShortTermStore(redisClient)
	memorySvc := memory.NewService(memoryRepo, shortTermStore)
	memoryHandler := memory.NewHandler(memorySvc)

	// Governance (Phase 5)
	quotaRepo := quota.NewRepository(pool)
	rateLimiter := quota.NewRateLimiter(redisClient)
	quotaSvc := quota.NewService(quotaRepo, rateLimiter, cfg.Governance)
	auditRepo := audit.NewRepository(pool)
	govHandler := governance.NewHandler(quotaSvc, auditRepo)

	// NATS publisher and consumer manager
	publisher := inats.NewPublisher(natsClient.JetStream())
	consumerMgr := inats.NewConsumerManager(natsClient.JetStream())

	// Audit consumer: NATS → audit_logs table
	auditConsumer := audit.NewConsumer(auditRepo, consumerMgr)

	// Orchestrator
	validator := orchestrator.NewValidator()
	orchRouter := orchestrator.NewRouter(agentRepo)
	orch := orchestrator.NewOrchestrator(publisher, consumerMgr, validator, orchRouter, quotaSvc)

	// XMPP handler and component
	xmppHandler := ixmpp.NewHandler(publisher)
	xmppComp, err := ixmpp.NewComponent(cfg.XMPP, xmppHandler)
	if err != nil {
		slog.Error("creating XMPP component", "error", err)
		os.Exit(1)
	}

	// Outbound relay: NATS → XMPP
	outboundRelay := ixmpp.NewOutboundRelay(xmppHandler, xmppComp.Sender(), consumerMgr)

	// Worker pool + gRPC server
	workerPool := worker.NewPool()
	workerRepo := worker.NewRepository(pool)
	grpcWorkerServer := worker.NewServer(workerPool, workerRepo)

	var grpcServerOpts []grpc.ServerOption
	if cfg.GRPC.WorkerAPIKey != "" {
		grpcServerOpts = append(grpcServerOpts,
			grpc.UnaryInterceptor(worker.UnaryAuthInterceptor(cfg.GRPC.WorkerAPIKey)),
			grpc.StreamInterceptor(worker.StreamAuthInterceptor(cfg.GRPC.WorkerAPIKey)),
		)
	}
	grpcSrv := grpc.NewServer(grpcServerOpts...)
	pb.RegisterWorkerServiceServer(grpcSrv, grpcWorkerServer)

	// Task dispatcher: NATS tasks → gRPC workers → outbound messages
	dispatcher := worker.NewDispatcher(
		workerPool, publisher, consumerMgr,
		agentSvc, workerRepo, memorySvc, quotaSvc, grpcWorkerServer.ResultChannel(),
		cfg.GRPC.TaskTimeoutSec,
	)

	// Auth rate limiter
	authRateLimiter := middleware.NewRateLimiter(redisClient, 20, 60)

	// Router
	router := api.NewRouter(pool, natsClient, api.RouterConfig{
		CORSAllowedOrigins: cfg.Server.CORSAllowedOrigins,
		AuthRateLimiter:    authRateLimiter.Middleware,
	}, api.HandlerSet{
		Register: authHandler.Register,
		Login:    authHandler.Login,
		Refresh:  authHandler.Refresh,
		Logout:   authHandler.Logout,

		CreateAgent:         agentHandler.Create,
		ListAgents:          agentHandler.List,
		GetAgent:            agentHandler.Get,
		UpdateAgent:         agentHandler.Update,
		DeleteAgent:         agentHandler.Delete,
		OwnershipMiddleware: agentHandler.OwnershipMiddleware,

		ListMemories:      memoryHandler.List,
		CreateMemory:      memoryHandler.Create,
		SearchMemories:    memoryHandler.Search,
		DeleteMemory:      memoryHandler.Delete,
		DeleteAllMemories: memoryHandler.DeleteAll,

		GetUserQuota:       govHandler.GetQuota,
		ListAuditLogs:      govHandler.ListAuditLogs,
		ListAgentAuditLogs: govHandler.ListAgentAuditLogs,

		AuthMiddleware: auth.Middleware(authSvc),

		WorkerPoolHealthy: func() bool { return workerPool.ConnectedCount() > 0 },
	})

	// Start background goroutines
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("starting XMPP component")
		if err := xmppComp.Start(ctx); err != nil {
			slog.Error("XMPP component error", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("starting orchestrator")
		if err := orch.Start(ctx); err != nil {
			slog.Error("orchestrator error", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("starting outbound relay")
		if err := outboundRelay.Start(ctx); err != nil {
			slog.Error("outbound relay error", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
		lis, err := net.Listen("tcp", addr)
		if err != nil {
			slog.Error("gRPC listen error", "error", err)
			return
		}
		slog.Info("starting gRPC server", "addr", addr)
		if err := grpcSrv.Serve(lis); err != nil {
			slog.Error("gRPC server error", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("starting task dispatcher")
		if err := dispatcher.Start(ctx); err != nil {
			slog.Error("task dispatcher error", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("starting audit consumer")
		if err := auditConsumer.Start(ctx); err != nil {
			slog.Error("audit consumer error", "error", err)
		}
	}()

	// Start HTTP server (blocks until shutdown signal)
	srv := server.New(cfg.Server, router)
	if err := srv.Start(); err != nil {
		slog.Error("server error", "error", err)
	}

	// Ordered shutdown
	slog.Info("initiating shutdown")
	cancel()

	grpcSrv.GracefulStop()

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("all goroutines stopped")
	case <-time.After(15 * time.Second):
		slog.Warn("shutdown timed out after 15s, forcing exit")
	}

	natsClient.Close()
	redisClient.Close()
	pool.Close()
	slog.Info("shutdown complete")
}

func setupLogger(cfg config.LogConfig) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{}
	switch cfg.Level {
	case "debug":
		opts.Level = slog.LevelDebug
	case "info":
		opts.Level = slog.LevelInfo
	case "warn":
		opts.Level = slog.LevelWarn
	case "error":
		opts.Level = slog.LevelError
	default:
		opts.Level = slog.LevelInfo
	}

	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}
