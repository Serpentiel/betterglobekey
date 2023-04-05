// Package eventhandler encapsulates logic to handle keyboard events.
package eventhandler

import (
	"time"

	"github.com/Serpentiel/betterglobekey/pkg/inputsource"
	"github.com/Serpentiel/betterglobekey/pkg/logger"
	"github.com/spf13/viper"
)

// newFnKeyHandler returns a new fnKeyHandler.
func newFnKeyHandler(v *viper.Viper, l logger.Logger) *fnKeyHandler {
	return &fnKeyHandler{
		l: l,

		doublePressMaximumDelay: v.GetInt("double_press.maximum_delay"),
		primaryInputSources:     v.GetStringSlice("input_sources.primary"),
		additionalInputSources:  v.GetStringSlice("input_sources.additional"),
	}
}

var _ handler = (*fnKeyHandler)(nil)

// fnKeyHandler is a handler for the fn key up event.
type fnKeyHandler struct {
	// l is the logger.
	l logger.Logger

	// doublePressMaximumDelay is the maximum delay between two presses of the fn key for them to be considered as a
	// double press.
	doublePressMaximumDelay int

	// primaryInputSources is a slice of the primary input sources.
	primaryInputSources []string

	// additionalInputSources is a slice of the additional input sources.
	additionalInputSources []string

	// doublePressable is a bool that indicates if the key is double pressable.
	doublePressable bool

	// doublePressed is a bool that indicates if the key is double pressed.
	doublePressed bool

	// currentInputSource is the current input source.
	currentInputSource string

	// previousInputSource is the previous input source.
	previousInputSource string
}

// getNextInputSource returns the next input source.
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

	h.l.Info("input source set", "from", h.previousInputSource, "to", inputSource)
}

// KeyUp is called when the key is released.
func (h *fnKeyHandler) KeyUp() handlerFunc {
	return func() {
		if len(h.additionalInputSources) > 0 {
			doublePressTicker := time.NewTicker(time.Duration(h.doublePressMaximumDelay) * time.Millisecond)

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
			h.l.Info("globe key pressed")

			h.setInputSource(h.primaryInputSources[h.getNextInputSource(&h.primaryInputSources)])
		} else {
			// TODO: This is not working as designed at the moment—this is supposed to open the original
			//  input source popup, but implementing it requires some reverse engineering.
			//  There is probably a function in macOS private API that can be used to open the popup.
			h.l.Info("globe key double pressed")

			h.setInputSource(h.previousInputSource)

			h.currentInputSource = inputsource.Current()

			h.setInputSource(h.additionalInputSources[h.getNextInputSource(&h.additionalInputSources)])
		}
	}
}
