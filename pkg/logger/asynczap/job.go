// Package asynczap is the package that provides the AsyncZap logger.
package asynczap

import (
	"github.com/Serpentiel/betterglobekey/pkg/logger"
	"go.uber.org/zap/zapcore"
)

// NewJob creates a new Job instance.
func NewJob(l zapcore.Level, m string, kv []any) *Job {
	return &Job{
		level:         l,
		msg:           m,
		keysAndValues: kv,
	}
}

// Job defines an object which is used to enqueue logging jobs for AsyncZap.
type Job struct {
	level         zapcore.Level
	msg           string
	keysAndValues []any
}

// work is a function which performs the logging work for a Job instance.
func (j *Job) work(l logger.Logger) {
	switch j.level {
	case zapcore.DPanicLevel:
		l.DPanic(j.msg, j.keysAndValues...)
	case zapcore.DebugLevel:
		l.Debug(j.msg, j.keysAndValues...)
	case zapcore.ErrorLevel:
		l.Error(j.msg, j.keysAndValues...)
	case zapcore.FatalLevel:
		l.Fatal(j.msg, j.keysAndValues...)
	case zapcore.InfoLevel:
		l.Info(j.msg, j.keysAndValues...)
	case zapcore.PanicLevel:
		l.Panic(j.msg, j.keysAndValues...)
	case zapcore.WarnLevel:
		l.Warn(j.msg, j.keysAndValues...)
	}
}
