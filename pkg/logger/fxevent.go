// Package logger is the package that provides logging functionality.
package logger

import (
	"strings"

	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

// FxeventLogger is a wrapper around the Logger interface that implements the fxevent.Logger interface.
type FxeventLogger struct {
	Logger Logger
}

// LogEvent is called by the fx library to log events.
// nolint:funlen
func (l *FxeventLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.Logger.Info("OnStart hook executing", "callee", e.FunctionName, "caller", e.CallerName)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.Logger.Info("OnStart hook failed", "callee", e.FunctionName, "caller", e.CallerName, "error", e.Err)
		} else {
			l.Logger.Info("OnStart hook executed",
				"callee", e.FunctionName,
				"caller", e.CallerName,
				"runtime", e.Runtime.String(),
			)
		}
	case *fxevent.OnStopExecuting:
		l.Logger.Info("OnStop hook executing", "callee", e.FunctionName, "caller", e.CallerName)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.Logger.Info("OnStop hook failed", "callee", e.FunctionName, "caller", e.CallerName, "error", e.Err)
		} else {
			l.Logger.Info("OnStop hook executed",
				zap.String("callee", e.FunctionName),
				zap.String("caller", e.CallerName),
				zap.String("runtime", e.Runtime.String()),
			)
		}
	case *fxevent.Supplied:
		l.Logger.Info("supplied", "type", e.TypeName, "error", e.Err)
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.Logger.Info("provided", "constructor", e.ConstructorName, "type", rtype)
		}

		if e.Err != nil {
			l.Logger.Error("error encountered while applying options", "error", e.Err)
		}
	case *fxevent.Invoking:
		// Do not log stack as it will make logs hard to read.
		l.Logger.Info("invoking", "function", e.FunctionName)
	case *fxevent.Invoked:
		if e.Err != nil {
			l.Logger.Error("invoke failed", "error", e.Err, "stack", e.Trace, "function", e.FunctionName)
		}
	case *fxevent.Stopping:
		l.Logger.Info("received signal", "signal", strings.ToUpper(e.Signal.String()))
	case *fxevent.Stopped:
		if e.Err != nil {
			l.Logger.Error("stop failed", "error", e.Err)
		}
	case *fxevent.RollingBack:
		l.Logger.Error("start failed, rolling back", "error", e.StartErr)
	case *fxevent.RolledBack:
		if e.Err != nil {
			l.Logger.Error("rollback failed", "error", e.Err)
		}
	case *fxevent.Started:
		if e.Err != nil {
			l.Logger.Error("start failed", "error", e.Err)
		} else {
			l.Logger.Info("started")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			l.Logger.Error("custom Logger initialization failed", "error", e.Err)
		} else {
			l.Logger.Info("initialized custom fxevent.Logger", "function", e.ConstructorName)
		}
	}
}
