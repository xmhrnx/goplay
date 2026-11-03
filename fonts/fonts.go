// Package fonts embeds the Noto fonts GoPlay falls back to for scripts
// the platform's UI font lacks: language names in their own script and
// the interface in each translation.
package fonts

import "embed"

// FS holds the font files, each a variable font at its default weight.
//
//go:embed *.ttf
var FS embed.FS
