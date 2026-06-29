// Package logging provides the application's zap-based logger, writing
// human-readable output to the console and structured JSON to a rotating file.
package logging

import (
	"os"

	"github.com/Serpentiel/betterglobekey/internal/domain/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is a zap-backed logger.
type Logger struct {
	sugar *zap.SugaredLogger
}

// New builds a Logger that tees human-readable console output and structured
// JSON to a rotating file, both filtered at the configured level.
func New(cfg config.Logger) *Logger {
	level := parseLevel(cfg.Level)

	console := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		zapcore.AddSync(os.Stdout),
		level,
	)

	file := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Path,
			MaxAge:     cfg.RetentionDays,
			MaxBackups: cfg.RetentionFiles,
			Compress:   true,
		}),
		level,
	)

	logger := zap.New(zapcore.NewTee(console, file), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{sugar: logger.Sugar()}
}

// parseLevel maps a configured level name to a zap level, defaulting to info.
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// Info logs an informational message with optional key-value context.
func (l *Logger) Info(msg string, keysAndValues ...any) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Error logs an error message with optional key-value context.
func (l *Logger) Error(msg string, keysAndValues ...any) {
	l.sugar.Errorw(msg, keysAndValues...)
}

// Close flushes any buffered log entries.
func (l *Logger) Close() error {
	// Syncing the console core can fail on some platforms (e.g. stdout is not a
	// syncable device); that error is not actionable, so it is ignored.
	_ = l.sugar.Sync()

	return nil
}
