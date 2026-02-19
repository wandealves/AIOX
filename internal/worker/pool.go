package worker

import (
	"sync"

	pb "github.com/aiox-platform/aiox/internal/worker/workerpb"
	"google.golang.org/grpc"
)

// ConnectedWorker represents a Python worker connected via gRPC bidirectional stream.
type ConnectedWorker struct {
	WorkerID           string
	MaxConcurrent      int32
	SupportedProviders []string

	mu          sync.Mutex
	ActiveTasks int32
	Stream      grpc.BidiStreamingServer[pb.WorkerMessage, pb.ServerMessage]
}

// Send safely sends a ServerMessage to the worker's stream.
func (w *ConnectedWorker) Send(msg *pb.ServerMessage) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.Stream.Send(msg)
}

// IncrementActive atomically increments the active task count.
func (w *ConnectedWorker) IncrementActive() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ActiveTasks++
}

// DecrementActive atomically decrements the active task count.
func (w *ConnectedWorker) DecrementActive() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.ActiveTasks > 0 {
		w.ActiveTasks--
	}
}

// LoadFraction returns ActiveTasks / MaxConcurrent as a float for load balancing.
func (w *ConnectedWorker) LoadFraction() float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.MaxConcurrent <= 0 {
		return 1.0
	}
	return float64(w.ActiveTasks) / float64(w.MaxConcurrent)
}

// Pool manages connected Python workers.
type Pool struct {
	mu      sync.RWMutex
	workers map[string]*ConnectedWorker
}

// NewPool creates a new worker pool.
func NewPool() *Pool {
	return &Pool{
		workers: make(map[string]*ConnectedWorker),
	}
}

// Register adds a worker to the pool. Returns false if a worker with the same ID is already connected.
func (p *Pool) Register(w *ConnectedWorker) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, exists := p.workers[w.WorkerID]; exists {
		return false
	}
	p.workers[w.WorkerID] = w
	return true
}

// Unregister removes a worker from the pool.
func (p *Pool) Unregister(workerID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.workers, workerID)
}

// SelectWorker picks the least-loaded worker that has capacity.
// Returns nil if no workers are available.
func (p *Pool) SelectWorker() *ConnectedWorker {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var best *ConnectedWorker
	bestLoad := float64(2.0) // > 1.0 means none found yet

	for _, w := range p.workers {
		load := w.LoadFraction()
		if load >= 1.0 {
			continue // fully loaded
		}
		if load < bestLoad {
			bestLoad = load
			best = w
		}
	}
	return best
}

// ConnectedCount returns the number of connected workers.
func (p *Pool) ConnectedCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.workers)
}

// Get returns a worker by ID, or nil if not found.
func (p *Pool) Get(workerID string) *ConnectedWorker {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.workers[workerID]
}
