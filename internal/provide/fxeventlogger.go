// Package provide is the package that contains all of the providers for the dependency injection container.
package provide

import (
	"context"

	"github.com/Serpentiel/betterglobekey/pkg/logger"
	"github.com/Serpentiel/betterglobekey/pkg/logger/asynczap"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// FxeventLogger is a function which provides an fxevent.Logger instance.
func FxeventLogger(lc fx.Lifecycle, l logger.Logger) fxevent.Logger {
	az, _ := l.(*asynczap.AsyncZap)

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			return az.Logger.Close()
		},
	})

	return &logger.FxeventLogger{Logger: az.Logger}
}
