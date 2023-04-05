// Package eventhandler encapsulates logic to handle keyboard events.
package eventhandler

import (
	"context"

	"github.com/Serpentiel/betterglobekey/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

// NewEventHandler is the constructor for the event handler.
func NewEventHandler(lc fx.Lifecycle, v *viper.Viper, l logger.Logger) (e *EventHandler) {
	e = &EventHandler{}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			e.handlers = HandlersMap{
				Rawcode["fn"]: newFnKeyHandler(v, l),
			}

			return nil
		},
	})

	return e
}

// EventHandler is a type that represents an event handler.
type EventHandler struct {
	// handlers is a map of rawcode to a map of kind to handler.
	handlers HandlersMap
}

// HandlersMap is a type for a map of handlers.
type HandlersMap map[rawCode]handler

// handlerFunc is a function that handles an event.
type handlerFunc func()

// handler is a type that represents a handler for an event.
type handler interface {
	// KeyUp is called when the key is released.
	KeyUp() handlerFunc
}

// Handlers returns the handlers.
func (e *EventHandler) Handlers() HandlersMap {
	return e.handlers
}
