// Package eventhandler encapsulates logic to handle keyboard events.
package eventhandler

// keyFn is the rawcode for the fn key.
const keyFn rawCode = 179

// rawCode is a type that holds the rawcode.
type rawCode = uint16

// Rawcode is a map of name to rawcode.
var Rawcode = map[string]rawCode{
	"fn": keyFn,
}
