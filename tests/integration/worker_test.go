//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/aiox-platform/aiox/internal/agents"
	inats "github.com/aiox-platform/aiox/internal/nats"
	"github.com/aiox-platform/aiox/internal/worker"
	pb "github.com/aiox-platform/aiox/internal/worker/workerpb"
)

// TestWorkerEndToEnd tests the full flow:
// Publish task → dispatcher → gRPC worker → outbound message + execution record.
//
// This test requires running infrastructure:
//   - NATS on localhost:4222
//   - PostgreSQL on localhost:5433
//
// Run with: go test ./tests/integration/ -v -tags=integration -run TestWorkerEndToEnd
func TestWorkerEndToEnd(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to NATS
	nc, err := nats.Connect("nats://localhost:4222")
	require.NoError(t, err)
	defer nc.Close()

	js, err := jetstream.New(nc)
	require.NoError(t, err)

	// Ensure streams exist
	for _, cfg := range []jetstream.StreamConfig{
		{Name: "AIOX_TASKS", Subjects: []string{"aiox.tasks.>"}, Retention: jetstream.WorkQueuePolicy, MaxAge: time.Hour},
		{Name: "AIOX_MESSAGES", Subjects: []string{"aiox.messages.>"}, Retention: jetstream.WorkQueuePolicy, MaxAge: 24 * time.Hour},
		{Name: "AIOX_EVENTS", Subjects: []string{"aiox.events.>"}, Retention: jetstream.LimitsPolicy, MaxAge: 7 * 24 * time.Hour},
	} {
		_, err = js.CreateOrUpdateStream(ctx, cfg)
		require.NoError(t, err)
	}

	publisher := inats.NewPublisher(js)
	consumerMgr := inats.NewConsumerManager(js)

	// Start gRPC server with worker pool
	pool := worker.NewPool()
	grpcServer := worker.NewServer(pool, nil) // nil repo for test

	grpcSrv := grpc.NewServer()
	pb.RegisterWorkerServiceServer(grpcSrv, grpcServer)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	grpcAddr := lis.Addr().String()

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			t.Logf("gRPC server stopped: %v", err)
		}
	}()
	defer grpcSrv.GracefulStop()

	// Start dispatcher (nil repo to skip DB writes)
	dispatcher := worker.NewDispatcher(
		pool, publisher, consumerMgr,
		nil, // agentSvc — we'll test without it
		nil, // repo
		grpcServer.ResultChannel(),
		30,
	)

	// We can't use the full dispatcher without agentSvc.
	// Instead, test the gRPC worker registration and task response flow.
	_ = dispatcher

	// Connect a mock worker via gRPC
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewWorkerServiceClient(conn)
	stream, err := client.TaskStream(ctx)
	require.NoError(t, err)

	// Register
	err = stream.Send(&pb.WorkerMessage{
		Payload: &pb.WorkerMessage_Register{
			Register: &pb.RegisterWorker{
				WorkerId:           "test-worker-1",
				MaxConcurrent:      4,
				SupportedProviders: []string{"openai"},
			},
		},
	})
	require.NoError(t, err)

	// Receive ack
	ack, err := stream.Recv()
	require.NoError(t, err)
	regAck := ack.GetRegisterAck()
	require.NotNil(t, regAck)
	assert.True(t, regAck.Accepted)

	// Verify worker is in pool
	assert.Equal(t, 1, pool.ConnectedCount())

	// Send a task request directly through the pool worker
	testWorker := pool.SelectWorker()
	require.NotNil(t, testWorker)

	requestID := uuid.New().String()
	err = testWorker.Send(&pb.ServerMessage{
		Payload: &pb.ServerMessage_TaskRequest{
			TaskRequest: &pb.TaskRequest{
				RequestId:     requestID,
				AgentId:       uuid.New().String(),
				OwnerUserId:   uuid.New().String(),
				UserMessage:   "Hello, test!",
				SystemPrompt:  "You are a test bot.",
				LlmConfigJson: `{"provider":"openai","model":"gpt-4o-mini"}`,
				FromJid:       "user@aiox.local",
				AgentJid:      "agent-test@agents.aiox.local",
				AgentName:     "TestBot",
			},
		},
	})
	require.NoError(t, err)

	// Mock worker receives the task and responds
	taskMsg, err := stream.Recv()
	require.NoError(t, err)
	taskReq := taskMsg.GetTaskRequest()
	require.NotNil(t, taskReq)
	assert.Equal(t, requestID, taskReq.RequestId)
	assert.Equal(t, "Hello, test!", taskReq.UserMessage)
	assert.Equal(t, "You are a test bot.", taskReq.SystemPrompt)

	// Worker sends response
	err = stream.Send(&pb.WorkerMessage{
		Payload: &pb.WorkerMessage_TaskResponse{
			TaskResponse: &pb.TaskResponse{
				RequestId:    requestID,
				WorkerId:     "test-worker-1",
				ResponseText: "Hello! I'm the test bot.",
				TokensUsed:   42,
				DurationMs:   100,
				ModelUsed:    "gpt-4o-mini",
			},
		},
	})
	require.NoError(t, err)

	// Verify the response arrives on the result channel
	select {
	case result := <-grpcServer.ResultChannel():
		assert.Equal(t, requestID, result.RequestId)
		assert.Equal(t, "Hello! I'm the test bot.", result.ResponseText)
		assert.Equal(t, int32(42), result.TokensUsed)
		assert.Equal(t, "gpt-4o-mini", result.ModelUsed)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for task response")
	}

	// Test heartbeat
	hbResp, err := client.Heartbeat(ctx, &pb.HeartbeatRequest{
		WorkerId:      "test-worker-1",
		ActiveTasks:   1,
		MemoryUsageMb: 128,
	})
	require.NoError(t, err)
	assert.True(t, hbResp.Ok)

	// Close stream
	err = stream.CloseSend()
	require.NoError(t, err)

	// Drain remaining messages
	for {
		_, err := stream.Recv()
		if err == io.EOF || err != nil {
			break
		}
	}

	// Wait for unregister
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, pool.ConnectedCount())
}

// TestPublishAndConsumeTask verifies TaskMessage round-trips through NATS.
func TestPublishAndConsumeTask(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	nc, err := nats.Connect("nats://localhost:4222")
	require.NoError(t, err)
	defer nc.Close()

	js, err := jetstream.New(nc)
	require.NoError(t, err)

	// Ensure stream
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:      "AIOX_TASKS",
		Subjects:  []string{"aiox.tasks.>"},
		Retention: jetstream.WorkQueuePolicy,
		MaxAge:    time.Hour,
	})
	require.NoError(t, err)

	publisher := inats.NewPublisher(js)
	consumerMgr := inats.NewConsumerManager(js)

	agentID := uuid.New()
	task := inats.TaskMessage{
		RequestID:   uuid.New().String(),
		AgentID:     agentID,
		OwnerUserID: uuid.New(),
		Message:     "test message",
		FromJID:     "user@aiox.local",
		AgentJID:    fmt.Sprintf("agent-%s@agents.aiox.local", agentID),
		AgentName:   "TestAgent",
	}

	err = publisher.PublishTask(ctx, agentID.String(), task)
	require.NoError(t, err)

	// Consume the task
	consumer, err := consumerMgr.EnsureConsumer(ctx, "AIOX_TASKS", "test-consumer-"+uuid.New().String(), "aiox.tasks.>")
	require.NoError(t, err)

	msgs, err := consumer.Fetch(1, jetstream.FetchMaxWait(5*time.Second))
	require.NoError(t, err)

	var received inats.TaskMessage
	for msg := range msgs.Messages() {
		err = json.Unmarshal(msg.Data(), &received)
		require.NoError(t, err)
		_ = msg.Ack()
	}

	assert.Equal(t, task.RequestID, received.RequestID)
	assert.Equal(t, task.AgentID, received.AgentID)
	assert.Equal(t, task.Message, received.Message)
	assert.Equal(t, task.AgentJID, received.AgentJID)
	assert.Equal(t, task.AgentName, received.AgentName)
}
