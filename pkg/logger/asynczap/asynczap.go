// Package asynczap is the package that provides the AsyncZap logger.
package asynczap

import (
	"github.com/Serpentiel/betterglobekey/pkg/logger"
	"go.uber.org/zap/zapcore"
)

// NewAsyncZap creates a new AsyncZap instance.
func NewAsyncZap(l logger.Logger) *AsyncZap {
	// queueSize is the size of the queue for the AsyncZap.
	const queueSize = 1000

	return &AsyncZap{
		Logger: l,

		jobs: make(chan *Job, queueSize),
		quit: make(chan struct{}),
	}
}

var _ logger.Logger = (*AsyncZap)(nil)

// AsyncZap defines an object that performs asynchronous logging and implements log.Logger.
type AsyncZap struct {
	Logger logger.Logger

	jobs chan *Job
	quit chan struct{}
}

// Work is a function which processes Job instances in the AsyncZap queue.
func (a *AsyncZap) Work() {
	for {
		select {
		case j := <-a.jobs:
			j.work(a.Logger)
		case <-a.quit:
			return
		}
	}
}

// Debug logs a message with some additional context.
func (a *AsyncZap) Debug(msg string, keysAndValues ...any) {
	a.jobs <- NewJob(zapcore.DebugLevel, msg, keysAndValues)
}

// DPanic logs a message with some additional context. In development mode, the logger then panics.
func (a *AsyncZap) DPanic(msg string, keysAndValues ...any) {
	a.jobs <- NewJob(zapcore.DPanicLevel, msg, keysAndValues)
}

// Info logs a message with some additional context.
func (a *AsyncZap) Info(msg string, keysAndValues ...any) {
	a.jobs <- NewJob(zapcore.InfoLevel, msg, keysAndValues)
}

// Warn logs a message with some additional context.
func (a *AsyncZap) Warn(msg string, keysAndValues ...any) {
	a.jobs <- NewJob(zapcore.WarnLevel, msg, keysAndValues)
}

// Error logs a message with some additional context.
func (a *AsyncZap) Error(msg string, keysAndValues ...any) {
	a.jobs <- NewJob(zapcore.ErrorLevel, msg, keysAndValues)
}

// Fatal logs a message with some additional context, then calls os.Exit.
func (a *AsyncZap) Fatal(msg string, keysAndValues ...any) {
	a.jobs <- NewJob(zapcore.FatalLevel, msg, keysAndValues)
}

// Panic logs a message with some additional context, then panics.
func (a *AsyncZap) Panic(msg string, keysAndValues ...any) {
	a.jobs <- NewJob(zapcore.PanicLevel, msg, keysAndValues)
}

// Close closes the logger.
func (a *AsyncZap) Close() (err error) {
	a.quit <- struct{}{}

	close(a.jobs)

	for j := range a.jobs {
		j.work(a.Logger)
	}

	return
}
