// Package assets is the package that contains all of the assets for the application.
package assets

import _ "embed" // embed is imported as a blank import to let us embed the assets.

// ExampleConfig is the example config for the application.
//
//go:embed .betterglobekey.example.yaml
var ExampleConfig []byte
