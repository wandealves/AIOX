package worker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_RegisterAndCount(t *testing.T) {
	pool := NewPool()
	assert.Equal(t, 0, pool.ConnectedCount())

	w := &ConnectedWorker{WorkerID: "w1", MaxConcurrent: 4}
	ok := pool.Register(w)
	require.True(t, ok)
	assert.Equal(t, 1, pool.ConnectedCount())
}

func TestPool_RegisterDuplicate(t *testing.T) {
	pool := NewPool()

	w1 := &ConnectedWorker{WorkerID: "w1", MaxConcurrent: 4}
	require.True(t, pool.Register(w1))

	w2 := &ConnectedWorker{WorkerID: "w1", MaxConcurrent: 8}
	assert.False(t, pool.Register(w2))
	assert.Equal(t, 1, pool.ConnectedCount())
}

func TestPool_Unregister(t *testing.T) {
	pool := NewPool()

	w := &ConnectedWorker{WorkerID: "w1", MaxConcurrent: 4}
	pool.Register(w)
	assert.Equal(t, 1, pool.ConnectedCount())

	pool.Unregister("w1")
	assert.Equal(t, 0, pool.ConnectedCount())
}

func TestPool_UnregisterNonexistent(t *testing.T) {
	pool := NewPool()
	pool.Unregister("nonexistent") // should not panic
	assert.Equal(t, 0, pool.ConnectedCount())
}

func TestPool_SelectWorker_LeastLoaded(t *testing.T) {
	pool := NewPool()

	w1 := &ConnectedWorker{WorkerID: "w1", MaxConcurrent: 4, ActiveTasks: 3}
	w2 := &ConnectedWorker{WorkerID: "w2", MaxConcurrent: 4, ActiveTasks: 1}
	w3 := &ConnectedWorker{WorkerID: "w3", MaxConcurrent: 4, ActiveTasks: 2}

	pool.Register(w1)
	pool.Register(w2)
	pool.Register(w3)

	selected := pool.SelectWorker()
	require.NotNil(t, selected)
	assert.Equal(t, "w2", selected.WorkerID, "should select least loaded worker")
}

func TestPool_SelectWorker_NoneAvailable(t *testing.T) {
	pool := NewPool()
	assert.Nil(t, pool.SelectWorker(), "empty pool should return nil")
}

func TestPool_SelectWorker_AllFullyLoaded(t *testing.T) {
	pool := NewPool()

	w1 := &ConnectedWorker{WorkerID: "w1", MaxConcurrent: 2, ActiveTasks: 2}
	w2 := &ConnectedWorker{WorkerID: "w2", MaxConcurrent: 3, ActiveTasks: 3}

	pool.Register(w1)
	pool.Register(w2)

	assert.Nil(t, pool.SelectWorker(), "all fully loaded should return nil")
}

func TestPool_Get(t *testing.T) {
	pool := NewPool()

	w := &ConnectedWorker{WorkerID: "w1", MaxConcurrent: 4}
	pool.Register(w)

	got := pool.Get("w1")
	assert.Equal(t, w, got)

	got = pool.Get("nonexistent")
	assert.Nil(t, got)
}

func TestConnectedWorker_IncrementDecrement(t *testing.T) {
	w := &ConnectedWorker{WorkerID: "w1", MaxConcurrent: 4}
	assert.Equal(t, int32(0), w.ActiveTasks)

	w.IncrementActive()
	assert.Equal(t, int32(1), w.ActiveTasks)

	w.IncrementActive()
	assert.Equal(t, int32(2), w.ActiveTasks)

	w.DecrementActive()
	assert.Equal(t, int32(1), w.ActiveTasks)

	w.DecrementActive()
	assert.Equal(t, int32(0), w.ActiveTasks)

	// Should not go negative
	w.DecrementActive()
	assert.Equal(t, int32(0), w.ActiveTasks)
}

func TestConnectedWorker_LoadFraction(t *testing.T) {
	w := &ConnectedWorker{WorkerID: "w1", MaxConcurrent: 4, ActiveTasks: 2}
	assert.InDelta(t, 0.5, w.LoadFraction(), 0.001)

	w.ActiveTasks = 0
	assert.InDelta(t, 0.0, w.LoadFraction(), 0.001)

	w.ActiveTasks = 4
	assert.InDelta(t, 1.0, w.LoadFraction(), 0.001)

	// Zero max concurrent → fully loaded
	w2 := &ConnectedWorker{WorkerID: "w2", MaxConcurrent: 0}
	assert.InDelta(t, 1.0, w2.LoadFraction(), 0.001)
}
