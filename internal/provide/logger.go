// Package provide is the package that contains all of the providers for the dependency injection container.
package provide

import (
	"context"

	"github.com/Serpentiel/betterglobekey/pkg/logger"
	"github.com/Serpentiel/betterglobekey/pkg/logger/asynczap"
	"github.com/Serpentiel/betterglobekey/pkg/logger/zap"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// Logger is a function which provides a log.Logger instance.
func Logger(lc fx.Lifecycle, v *viper.Viper) logger.Logger {
	l := asynczap.NewAsyncZap(zap.NewZap(
		&zap.Params{
			Path:         v.GetString("logger.path"),
			RetainDays:   v.GetInt("logger.retain.days"),
			RetainCopies: v.GetInt("logger.retain.copies"),
		},
	))

	lc.Append(fx.Hook{
		OnStart: func(context.Context) (err error) {
			go l.Work()

			return
		},
		OnStop: func(context.Context) error {
			return l.Close()
		},
	})

	return l
}
