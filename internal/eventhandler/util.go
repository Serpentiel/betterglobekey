// Package eventhandler encapsulates logic to handle keyboard events.
package eventhandler

import (
	"fmt"

	"github.com/spf13/viper"
)

// intGet returns the value of the key as int.
func intGet(k string) int {
	result, ok := viper.Get(k).(int)
	if !ok {
		logger.Fatalw("config key type is not int", "key", k)
	}

	return result
}

// interfaceSliceGet returns the value of the key as []interface{}.
func interfaceSliceGet(k string) []interface{} {
	result, ok := viper.Get(k).([]interface{})
	if !ok {
		logger.Fatalw("config key type is not []interface{}", "key", k)
	}

	return result
}

// stringSlicesGet returns the value of the key as []string.
func stringSliceGet(k string) []string {
	val := interfaceSliceGet(k)

	result := make([]string, len(val))

	for i, v := range val {
		cv, ok := v.(string)
		if !ok {
			logger.Fatalw("config key type is not string", "key", fmt.Sprintf("%s.%d", k, i))
		}

		result[i] = cv
	}

	return result
}
