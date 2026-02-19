package backends

import (
	"bytes"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// mockBackend is a simple Backend for testing that records calls.
type mockBackend struct {
	putCalled   atomic.Int64
	getCalled   atomic.Int64
	hasCalled   atomic.Int64
	touchCalled atomic.Int64
	clearCalled atomic.Int64
	closeCalled atomic.Int64

	// Return values for Get
	getOutputID []byte
	getBody     []byte
	getSize     int64
	getPutTime  *time.Time
	getMiss     bool
}

func (m *mockBackend) Put(actionID, outputID []byte, body io.Reader, bodySize int64) error {
	m.putCalled.Add(1)
	return nil
}

func (m *mockBackend) Has(actionID []byte) (bool, error) {
	m.hasCalled.Add(1)
	return true, nil
}

func (m *mockBackend) Get(actionID []byte) ([]byte, io.ReadCloser, int64, *time.Time, bool, error) {
	m.getCalled.Add(1)
	if m.getMiss {
		return nil, nil, 0, nil, true, nil
	}
	return m.getOutputID, io.NopCloser(bytes.NewReader(m.getBody)), m.getSize, m.getPutTime, false, nil
}

func (m *mockBackend) Touch(actionID []byte) error {
	m.touchCalled.Add(1)
	return nil
}

func (m *mockBackend) Clear() error {
	m.clearCalled.Add(1)
	return nil
}

func (m *mockBackend) Close() error {
	m.closeCalled.Add(1)
	return nil
}

func TestReadOnly_PutIsNoOp(t *testing.T) {
	inner := &mockBackend{}
	ro := NewReadOnly(inner)

	body := strings.NewReader("hello world")
	err := ro.Put([]byte("action"), []byte("output"), body, 11)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if inner.putCalled.Load() != 0 {
		t.Fatalf("expected inner Put not to be called, but it was called %d times", inner.putCalled.Load())
	}

	stats := ro.Stats()
	if stats.PutsSkipped != 1 {
		t.Fatalf("expected PutsSkipped=1, got %d", stats.PutsSkipped)
	}
}

func TestReadOnly_PutDrainsBody(t *testing.T) {
	inner := &mockBackend{}
	ro := NewReadOnly(inner)

	pr, pw := io.Pipe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err := pw.Write([]byte("test data"))
		if err != nil {
			t.Errorf("unexpected write error: %v", err)
			return
		}
		pw.Close()
	}()

	err := ro.Put([]byte("action"), []byte("output"), pr, 9)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	<-done // Ensures the writer goroutine completes without blocking
}

func TestReadOnly_PutNilBody(t *testing.T) {
	inner := &mockBackend{}
	ro := NewReadOnly(inner)

	err := ro.Put([]byte("action"), []byte("output"), nil, 0)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestReadOnly_TouchIsNoOp(t *testing.T) {
	inner := &mockBackend{}
	ro := NewReadOnly(inner)

	err := ro.Touch([]byte("action"))
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if inner.touchCalled.Load() != 0 {
		t.Fatalf("expected inner Touch not to be called, but it was called %d times", inner.touchCalled.Load())
	}

	stats := ro.Stats()
	if stats.TouchesSkipped != 1 {
		t.Fatalf("expected TouchesSkipped=1, got %d", stats.TouchesSkipped)
	}
}

func TestReadOnly_GetPassesThrough(t *testing.T) {
	now := time.Now()
	inner := &mockBackend{
		getOutputID: []byte("out123"),
		getBody:     []byte("cached data"),
		getSize:     11,
		getPutTime:  &now,
		getMiss:     false,
	}
	ro := NewReadOnly(inner)

	outputID, body, size, putTime, miss, err := ro.Get([]byte("action"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if miss {
		t.Fatal("expected hit, got miss")
	}
	if string(outputID) != "out123" {
		t.Fatalf("expected outputID=out123, got %s", outputID)
	}
	if size != 11 {
		t.Fatalf("expected size=11, got %d", size)
	}
	if putTime == nil || !putTime.Equal(now) {
		t.Fatalf("expected putTime=%v, got %v", now, putTime)
	}

	data, _ := io.ReadAll(body)
	body.Close()
	if string(data) != "cached data" {
		t.Fatalf("expected body='cached data', got '%s'", data)
	}

	if inner.getCalled.Load() != 1 {
		t.Fatalf("expected inner Get called once, got %d", inner.getCalled.Load())
	}
}

func TestReadOnly_HasPassesThrough(t *testing.T) {
	inner := &mockBackend{}
	ro := NewReadOnly(inner)

	exists, err := ro.Has([]byte("action"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Fatal("expected exists=true")
	}
	if inner.hasCalled.Load() != 1 {
		t.Fatalf("expected inner Has called once, got %d", inner.hasCalled.Load())
	}
}

func TestReadOnly_ClearReturnsError(t *testing.T) {
	inner := &mockBackend{}
	ro := NewReadOnly(inner)

	err := ro.Clear()
	if err == nil {
		t.Fatal("expected error from Clear in read-only mode")
	}
	if !strings.Contains(err.Error(), "read-only") {
		t.Fatalf("expected error message to mention 'read-only', got: %v", err)
	}

	if inner.clearCalled.Load() != 0 {
		t.Fatal("expected inner Clear not to be called")
	}

	stats := ro.Stats()
	if stats.ClearsBlocked != 1 {
		t.Fatalf("expected ClearsBlocked=1, got %d", stats.ClearsBlocked)
	}
}

func TestReadOnly_ClosePassesThrough(t *testing.T) {
	inner := &mockBackend{}
	ro := NewReadOnly(inner)

	err := ro.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.closeCalled.Load() != 1 {
		t.Fatalf("expected inner Close called once, got %d", inner.closeCalled.Load())
	}
}

func TestReadOnly_Unwrap(t *testing.T) {
	inner := &mockBackend{}
	ro := NewReadOnly(inner)

	unwrapped := ro.Unwrap()
	if unwrapped != inner {
		t.Fatal("expected Unwrap to return the inner backend")
	}
}

func TestReadOnly_StatsAccumulate(t *testing.T) {
	inner := &mockBackend{}
	ro := NewReadOnly(inner)

	for i := 0; i < 5; i++ {
		_ = ro.Put([]byte("a"), []byte("b"), nil, 0)
	}
	for i := 0; i < 3; i++ {
		_ = ro.Touch([]byte("a"))
	}
	for i := 0; i < 2; i++ {
		_ = ro.Clear()
	}

	stats := ro.Stats()
	if stats.PutsSkipped != 5 {
		t.Fatalf("expected PutsSkipped=5, got %d", stats.PutsSkipped)
	}
	if stats.TouchesSkipped != 3 {
		t.Fatalf("expected TouchesSkipped=3, got %d", stats.TouchesSkipped)
	}
	if stats.ClearsBlocked != 2 {
		t.Fatalf("expected ClearsBlocked=2, got %d", stats.ClearsBlocked)
	}
}
