// Package eventlistener encapsulates logic to listen for keyboard events.
package eventlistener

import (
	"github.com/Serpentiel/betterglobekey/internal/pkg/eventhandler"
	hook "github.com/robotn/gohook"
	"go.uber.org/zap"
)

// Run starts the eventlistener.
func Run(logger *zap.SugaredLogger) {
	events := hook.Start()

	defer hook.End()

	logger.Info("eventlistener has been initialized")

	handlers := eventhandler.Setup(logger)

	logger.Info("eventhandler has been initialized")

	for e := range events {
		handler, ok := handlers[e.Rawcode]
		if !ok {
			continue
		}

		if e.Kind == hook.KeyUp {
			handler.KeyUp()()
		}
	}
}
