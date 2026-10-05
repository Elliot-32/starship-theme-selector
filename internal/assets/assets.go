package assets

import _ "embed"

// DefaultSymbols is the bundled Starship [os.symbols] override.
//
//go:embed os-symbols.toml
var DefaultSymbols []byte
