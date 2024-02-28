// Package eventhandler encapsulates logic to handle keyboard events.
package eventhandler

import (
	"errors"
	"time"

	"github.com/Serpentiel/betterglobekey/pkg/inputsource"
	"github.com/Serpentiel/betterglobekey/pkg/logger"
	"github.com/Serpentiel/betterglobekey/pkg/util"
	"github.com/spf13/viper"
)

// errNoInputSourceAvailable is the error returned when no input source is available.
var errNoInputSourceAvailable = errors.New("no input source available")

// newFnKeyHandler initializes and returns a new fnKeyHandler instance.
func newFnKeyHandler(v *viper.Viper, l logger.Logger) *fnKeyHandler {
	inputSources := v.GetStringMapStringSlice("input_sources")

	currentInputSource := inputsource.Current()

	var currentCollection string

	for collection, sources := range inputSources {
		if util.Contains(sources, currentInputSource) {
			currentCollection = collection

			break
		}
	}

	if currentCollection == "" && len(inputSources) > 0 {
		for k := range inputSources {
			currentCollection = k

			break
		}
	}

	return &fnKeyHandler{
		l:                       l,
		doublePressMaximumDelay: v.GetInt("double_press.maximum_delay"),
		inputSources:            inputSources,

		lastInputSourceInCollection: make(map[string]string),
		currentCollection:           currentCollection,
		currentInputSource:          currentInputSource,
		lastPressTime:               time.Now(),
	}
}

// fnKeyHandler is a type that represents a handler for the fn key.
type fnKeyHandler struct {
	// l is the logger.
	l logger.Logger
	// doublePressMaximumDelay is the maximum delay between two presses to be considered a double press.
	doublePressMaximumDelay int
	// inputSources is a map of input source collections.
	inputSources map[string][]string

	// lastInputSourceInCollection is a map of the last input source used in each collection.
	lastInputSourceInCollection map[string]string
	// currentCollection is the current input source collection.
	currentCollection string
	// currentInputSource is the current input source.
	currentInputSource string
	// lastPressTime is the time of the last press.
	lastPressTime time.Time
}

// KeyUp handles key up events and determines whether it's a single or double press.
func (h *fnKeyHandler) KeyUp() handlerFunc {
	return func() {
		currentTime := time.Now()
		elapsed := currentTime.Sub(h.lastPressTime)
		h.lastPressTime = currentTime

		if elapsed <= time.Duration(h.doublePressMaximumDelay)*time.Millisecond {
			h.handleDoublePress()
		} else {
			h.handleSinglePress()
		}
	}
}

// handleSinglePress handles a single press of the fn key.
func (h *fnKeyHandler) handleSinglePress() {
	h.l.Info("globe key pressed")

	nextSource, err := h.getNextInputSource()
	if err != nil {
		h.l.Error("failed to get next input source", "error", err)

		return
	}

	h.lastInputSourceInCollection[h.currentCollection] = nextSource
	h.setInputSource(nextSource)
}

// handleDoublePress handles a double press of the fn key.
func (h *fnKeyHandler) handleDoublePress() {
	h.l.Info("globe key double pressed")

	h.decrementPreviousInputSource(h.currentCollection)

	h.switchCollection()
	h.setInputSourceToLastOrFirst()
}

// getNextInputSource returns the next input source within the current collection.
func (h *fnKeyHandler) getNextInputSource() (string, error) {
	collection, exists := h.inputSources[h.currentCollection]
	if !exists || len(collection) == 0 {
		return "", errNoInputSourceAvailable
	}

	for i, source := range collection {
		if source == h.currentInputSource {
			return collection[(i+1)%len(collection)], nil
		}
	}

	return collection[0], nil
}

// switchCollection cycles to the next input source collection.
func (h *fnKeyHandler) switchCollection() {
	var nextCollection string

	foundCurrent := false

	for name := range h.inputSources {
		if foundCurrent {
			nextCollection = name

			break
		}

		if name == h.currentCollection {
			foundCurrent = true
		}
	}

	if nextCollection == "" {
		for name := range h.inputSources {
			nextCollection = name

			break
		}
	}

	h.currentCollection = nextCollection
}

// setInputSourceToLastOrFirst sets the input source to the last used in the current collection or the first one.
func (h *fnKeyHandler) setInputSourceToLastOrFirst() {
	nextSource := h.lastInputSourceInCollection[h.currentCollection]

	if nextSource == "" {
		nextSource = h.inputSources[h.currentCollection][0]
	}

	if nextSource != "" {
		h.setInputSource(nextSource)
	}
}

// decrementPreviousInputSource decrements the last input source in the specified collection.
func (h *fnKeyHandler) decrementPreviousInputSource(collection string) {
	lastSource := h.lastInputSourceInCollection[collection]
	collectionSources := h.inputSources[collection]

	index := -1

	for i, source := range collectionSources {
		if source == lastSource {
			index = i

			break
		}
	}

	if index > 0 {
		h.lastInputSourceInCollection[collection] = collectionSources[index-1]
	} else if index == 0 {
		h.lastInputSourceInCollection[collection] = collectionSources[len(collectionSources)-1]
	}
}

// setInputSource sets the input source if it exists.
func (h *fnKeyHandler) setInputSource(inputSource string) {
	h.currentInputSource = inputSource

	if util.Contains(inputsource.All(), inputSource) {
		inputsource.Select(inputSource)

		h.l.Info("input source set", "source", inputSource)
	} else {
		h.l.Info("input source not found", "source", inputSource)
	}
}
