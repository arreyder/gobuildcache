package backends

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

// ReadOnly wraps a Backend and suppresses all write operations (Put, Touch, Clear)
// while allowing reads (Get, Has) to pass through. This is useful for CI workers
// (e.g., PR builds) that should consume the shared S3 cache without polluting it.
// The local disk cache continues to operate with full read-write access.
type ReadOnly struct {
	backend Backend

	putsSkipped    atomic.Int64
	touchesSkipped atomic.Int64
	clearsBlocked  atomic.Int64
}

// NewReadOnly creates a new read-only wrapper around an existing backend.
func NewReadOnly(backend Backend) *ReadOnly {
	return &ReadOnly{
		backend: backend,
	}
}

// Unwrap returns the underlying backend.
func (ro *ReadOnly) Unwrap() Backend {
	return ro.backend
}

// Get delegates to the inner backend (reads are allowed).
func (ro *ReadOnly) Get(actionID []byte) ([]byte, io.ReadCloser, int64, *time.Time, bool, error) {
	return ro.backend.Get(actionID)
}

// Has delegates to the inner backend (reads are allowed).
func (ro *ReadOnly) Has(actionID []byte) (bool, error) {
	return ro.backend.Has(actionID)
}

// Put is a no-op in read-only mode. It drains the body reader for safety and returns nil.
func (ro *ReadOnly) Put(actionID, outputID []byte, body io.Reader, bodySize int64) error {
	ro.putsSkipped.Add(1)
	// Drain the body so callers that expect it to be consumed don't hang.
	if body != nil {
		_, _ = io.Copy(io.Discard, body)
	}
	return nil
}

// Touch is a no-op in read-only mode.
func (ro *ReadOnly) Touch(actionID []byte) error {
	ro.touchesSkipped.Add(1)
	return nil
}

// Clear returns an error because it is a destructive operation that should not be
// silently ignored. Unlike Put (called implicitly by the Go compiler), Clear is
// only invoked by explicit user commands.
func (ro *ReadOnly) Clear() error {
	ro.clearsBlocked.Add(1)
	return fmt.Errorf("clear blocked: backend is in read-only mode")
}

// Close delegates to the inner backend.
func (ro *ReadOnly) Close() error {
	return ro.backend.Close()
}

// ReadOnlyStats holds statistics for the read-only wrapper.
type ReadOnlyStats struct {
	PutsSkipped    int64
	TouchesSkipped int64
	ClearsBlocked  int64
}

// Stats returns counters for skipped write operations.
func (ro *ReadOnly) Stats() ReadOnlyStats {
	return ReadOnlyStats{
		PutsSkipped:    ro.putsSkipped.Load(),
		TouchesSkipped: ro.touchesSkipped.Load(),
		ClearsBlocked:  ro.clearsBlocked.Load(),
	}
}
