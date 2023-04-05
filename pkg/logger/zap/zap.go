// Package zap is the package that provides the Zap logger.
package zap

import (
	"os"

	"github.com/Serpentiel/betterglobekey/pkg/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NewZap creates a new Zap instance.
func NewZap(c *Params) *Zap {
	t := zapcore.NewTee(
		newConsoleZapCore(),
		newFileZapCore(c),
	)

	z := zap.New(
		t,
		zap.AddStacktrace(
			zap.LevelEnablerFunc(func(l zapcore.Level) bool {
				return l >= zapcore.ErrorLevel
			}),
		),
	)

	return &Zap{logger: z.Sugar()}
}

// Params defines the parameters for the Zap logger.
type Params struct {
	// Path is the path to the log file.
	Path string

	// RetainDays is the number of days to retain log files.
	RetainDays int

	// RetainCopies is the number of log files to retain.
	RetainCopies int
}

func newConsoleZapCore() zapcore.Core {
	e := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	o := zapcore.AddSync(os.Stdout)

	return zapcore.NewCore(e, o, zapcore.DebugLevel)
}

func newFileZapCore(c *Params) zapcore.Core {
	e := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	o := zapcore.AddSync(
		&lumberjack.Logger{
			Filename:   c.Path,
			MaxAge:     c.RetainDays,
			MaxBackups: c.RetainCopies,
			Compress:   true,
		},
	)

	return zapcore.NewCore(e, o, zapcore.InfoLevel)
}

var _ logger.Logger = (*Zap)(nil)

// Zap defines an object that performs logging and implements log.Logger.
type Zap struct {
	logger *zap.SugaredLogger
}

// Debug logs a message with some additional context.
func (z *Zap) Debug(msg string, keysAndValues ...any) {
	z.logger.Debugw(msg, keysAndValues...)
}

// DPanic logs a message with some additional context. In development mode, the logger then panics.
func (z *Zap) DPanic(msg string, keysAndValues ...any) {
	z.logger.DPanicw(msg, keysAndValues...)
}

// Info logs a message with some additional context.
func (z *Zap) Info(msg string, keysAndValues ...any) {
	z.logger.Infow(msg, keysAndValues...)
}

// Warn logs a message with some additional context.
func (z *Zap) Warn(msg string, keysAndValues ...any) {
	z.logger.Warnw(msg, keysAndValues...)
}

// Error logs a message with some additional context.
func (z *Zap) Error(msg string, keysAndValues ...any) {
	z.logger.Errorw(msg, keysAndValues...)
}

// Fatal logs a message with some additional context, then calls os.Exit.
func (z *Zap) Fatal(msg string, keysAndValues ...any) {
	z.logger.Fatalw(msg, keysAndValues...)
}

// Panic logs a message with some additional context, then panics.
func (z *Zap) Panic(msg string, keysAndValues ...any) {
	z.logger.Panicw(msg, keysAndValues...)
}

// Close closes the logger.
func (z *Zap) Close() (err error) {
	_ = z.logger.Sync()

	return
}
