// Package logger is the package that provides logging functionality.
package logger

// Logger is the interface that wraps the basic logging methods.
type Logger interface {
	// Debug logs a message with some additional context.
	Debug(string, ...any)

	// DPanic logs a message with some additional context. In development mode, the logger then panics.
	DPanic(string, ...any)

	// Info logs a message with some additional context.
	Info(string, ...any)

	// Warn logs a message with some additional context.
	Warn(string, ...any)

	// Error logs a message with some additional context.
	Error(string, ...any)

	// Fatal logs a message with some additional context, then calls os.Exit.
	Fatal(string, ...any)

	// Panic logs a message with some additional context, then panics.
	Panic(string, ...any)

	// Close closes the logger.
	Close() error
}
