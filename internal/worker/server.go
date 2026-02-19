package worker

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"

	pb "github.com/aiox-platform/aiox/internal/worker/workerpb"
	"google.golang.org/grpc"
)

// Server implements the WorkerServiceServer gRPC interface.
type Server struct {
	pb.UnimplementedWorkerServiceServer

	pool     *Pool
	repo     *Repository
	resultCh chan *pb.TaskResponse
}

// NewServer creates a new gRPC worker server.
func NewServer(pool *Pool, repo *Repository) *Server {
	return &Server{
		pool:     pool,
		repo:     repo,
		resultCh: make(chan *pb.TaskResponse, 256),
	}
}

// ResultChannel returns the channel that receives task responses from workers.
func (s *Server) ResultChannel() <-chan *pb.TaskResponse {
	return s.resultCh
}

// TaskStream implements the bidirectional streaming RPC.
// First message from worker must be RegisterWorker.
// Subsequent messages are TaskResponse results.
func (s *Server) TaskStream(stream grpc.BidiStreamingServer[pb.WorkerMessage, pb.ServerMessage]) error {
	// First message must be registration
	firstMsg, err := stream.Recv()
	if err != nil {
		return err
	}

	reg := firstMsg.GetRegister()
	if reg == nil {
		slog.Warn("worker stream: first message was not RegisterWorker")
		return stream.Send(&pb.ServerMessage{
			Payload: &pb.ServerMessage_RegisterAck{
				RegisterAck: &pb.RegisterAck{
					Accepted: false,
					Message:  "first message must be RegisterWorker",
				},
			},
		})
	}

	maxConcurrent := reg.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 4
	}

	worker := &ConnectedWorker{
		WorkerID:           reg.WorkerId,
		MaxConcurrent:      maxConcurrent,
		SupportedProviders: reg.SupportedProviders,
		Stream:             stream,
	}

	if !s.pool.Register(worker) {
		slog.Warn("worker already registered", "worker_id", reg.WorkerId)
		return stream.Send(&pb.ServerMessage{
			Payload: &pb.ServerMessage_RegisterAck{
				RegisterAck: &pb.RegisterAck{
					Accepted: false,
					Message:  "worker_id already registered",
				},
			},
		})
	}

	slog.Info("worker registered",
		"worker_id", reg.WorkerId,
		"max_concurrent", maxConcurrent,
		"providers", reg.SupportedProviders,
	)

	// Upsert worker in DB
	caps, _ := json.Marshal(map[string]any{
		"providers":      reg.SupportedProviders,
		"max_concurrent": maxConcurrent,
	})
	if s.repo != nil {
		if err := s.repo.UpsertWorker(stream.Context(), reg.WorkerId, "grpc-stream", 0, caps); err != nil {
			slog.Error("upserting worker in DB", "error", err)
		}
	}

	// Send ack
	if err := stream.Send(&pb.ServerMessage{
		Payload: &pb.ServerMessage_RegisterAck{
			RegisterAck: &pb.RegisterAck{
				Accepted: true,
				Message:  "registered",
			},
		},
	}); err != nil {
		s.pool.Unregister(reg.WorkerId)
		return err
	}

	// Receive loop: read TaskResponse messages from the worker
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				slog.Info("worker disconnected (EOF)", "worker_id", reg.WorkerId)
			} else {
				slog.Warn("worker stream error", "worker_id", reg.WorkerId, "error", err)
			}
			break
		}

		resp := msg.GetTaskResponse()
		if resp == nil {
			slog.Debug("ignoring non-TaskResponse message from worker", "worker_id", reg.WorkerId)
			continue
		}

		resp.WorkerId = reg.WorkerId
		s.resultCh <- resp
	}

	// Cleanup on disconnect.
	// Use context.Background() because stream.Context() is already cancelled
	// by the time we reach here.
	s.pool.Unregister(reg.WorkerId)
	if s.repo != nil {
		if err := s.repo.MarkWorkerOffline(context.Background(), reg.WorkerId); err != nil {
			slog.Error("marking worker offline", "error", err)
		}
	}
	slog.Info("worker unregistered", "worker_id", reg.WorkerId)

	return nil
}

// Heartbeat handles periodic health pings from workers.
func (s *Server) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if s.repo != nil {
		if err := s.repo.UpdateWorkerHeartbeat(ctx, req.WorkerId, int(req.ActiveTasks), int(req.AvgLatencyMs), int(req.MemoryUsageMb)); err != nil {
			slog.Error("updating heartbeat", "error", err)
		}
	}

	return &pb.HeartbeatResponse{Ok: true}, nil
}
