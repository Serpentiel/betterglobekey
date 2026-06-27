package app

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakePresser struct{}

func (fakePresser) Press(bool) {}

type fakeKeyboard struct {
	started chan struct{}
	stop    chan struct{}

	mu      sync.Mutex
	stopped bool
}

func (k *fakeKeyboard) Listen(func(reverse bool)) {
	close(k.started)
	<-k.stop
}

func (k *fakeKeyboard) Stop() {
	k.mu.Lock()
	k.stopped = true
	k.mu.Unlock()
	close(k.stop)
}

type fakeLogger struct {
	mu     sync.Mutex
	closed bool
}

func (*fakeLogger) Info(string, ...any)  {}
func (*fakeLogger) Error(string, ...any) {}

func (l *fakeLogger) Close() error {
	l.mu.Lock()
	l.closed = true
	l.mu.Unlock()

	return nil
}

func TestDaemonRunStartsListenerAndShutsDownCleanly(t *testing.T) {
	keyboard := &fakeKeyboard{started: make(chan struct{}), stop: make(chan struct{})}
	logger := &fakeLogger{}
	daemon := NewDaemon(fakePresser{}, keyboard, logger)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() { done <- daemon.Run(ctx) }()

	select {
	case <-keyboard.started:
	case <-time.After(time.Second):
		t.Fatal("listener was not started")
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not return after context cancellation")
	}

	keyboard.mu.Lock()
	stopped := keyboard.stopped
	keyboard.mu.Unlock()

	if !stopped {
		t.Error("keyboard was not stopped")
	}

	logger.mu.Lock()
	closed := logger.closed
	logger.mu.Unlock()

	if !closed {
		t.Error("logger was not closed")
	}
}
