// Package eventhandler encapsulates logic to handle keyboard events.
package eventhandler

import (
	"time"

	"github.com/Serpentiel/betterglobekey/pkg/inputsource"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// logger is the logger for the application.
var logger *zap.SugaredLogger

// doublePressThreshold is the threshold for a double press.
const doublePressThreshold = 250

// handlers is a map of rawcode to a map of kind to handler.
var handlers HandlersMap

// HandlersMap is a type for a map of handlers.
type HandlersMap map[rawCode]handler

// handlerFunc is a function that handles an event.
type handlerFunc func()

// handler is a type that represents a handler for an event.
type handler interface {
	// KeyUp is called when the key is released.
	KeyUp() handlerFunc
}

// fnKeyHandler is a handler for the Fn key up event.
type fnKeyHandler struct {
	// doublePressable is a bool that indicates if the key is double pressable.
	doublePressable bool

	// doublePressed is a bool that indicates if the key is double pressed.
	doublePressed bool

	// inputSources is a slice of all available input sources.
	inputSources []string

	// currentInputSource is the current input source.
	currentInputSource string

	// previousInputSource is the previous input source.
	previousInputSource string

	// primaryInputSources is a slice of the primary input sources.
	primaryInputSources []string

	// additionalInputSources is a slice of the additional input sources.
	additionalInputSources []string
}

// newFnKeyHandler returns a new fnKeyHandler.
func newFnKeyHandler() *fnKeyHandler {
	return &fnKeyHandler{
		inputSources: inputsource.All(),

		// TODO: Make this configurable.
		primaryInputSources: viper.GetStringSlice("primary_input_sources"),

		// TODO: Make this configurable.
		additionalInputSources: []string{
			"com.apple.keylayout.Finnish",
		},
	}
}

func (h *fnKeyHandler) getNextInputSource(inputSources *[]string) int {
	nextInputSource := 0

	for k, v := range *inputSources {
		if v == h.currentInputSource {
			if k == len(*inputSources)-1 {
				break
			}

			nextInputSource = k + 1
		}
	}

	return nextInputSource
}

// setInputSource sets the input source.
func (h *fnKeyHandler) setInputSource(inputSource string) {
	h.previousInputSource = h.currentInputSource

	inputsource.Select(inputSource)

	logger.Infow("input source set", "from", h.previousInputSource, "to", inputSource)
}

// KeyUp is called when the key is released.
func (h *fnKeyHandler) KeyUp() handlerFunc {
	return func() {
		{
			doublePressTicker := time.NewTicker(doublePressThreshold * time.Millisecond)

			h.doublePressed = h.doublePressable

			h.doublePressable = !h.doublePressable

			go func() {
				if <-doublePressTicker.C; true {
					h.doublePressable = false

					doublePressTicker.Stop()

					return
				}
			}()
		}

		h.currentInputSource = inputsource.Current()

		if !h.doublePressed {
			logger.Info("globe key pressed")

			h.setInputSource(h.primaryInputSources[h.getNextInputSource(&h.primaryInputSources)])
		} else {
			// TODO: This isn't working as designed at the moment—this is supposed to open the original
			//  input source popup, but implementing it requires some reverse engineering.
			//  There is probably a function in macOS private API that can be used to open the popup.

			logger.Info("globe key double pressed")

			h.setInputSource(h.previousInputSource)

			h.currentInputSource = inputsource.Current()

			h.setInputSource(h.additionalInputSources[h.getNextInputSource(&h.additionalInputSources)])
		}
	}
}

// setup sets up package's dependencies.
func setup(l *zap.SugaredLogger) {
	logger = l
}

// Setup sets up the handlers.
func Setup(logger *zap.SugaredLogger) HandlersMap {
	setup(logger)

	handlers = HandlersMap{
		Rawcode["fn"]: newFnKeyHandler(),
	}

	return handlers
}
