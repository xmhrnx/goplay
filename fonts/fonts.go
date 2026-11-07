// Package fonts embeds the Noto fonts GoPlay falls back to for scripts
// the platform's UI font lacks: language names in their own script and
// the interface in each translation.
package fonts

import "embed"

// FS holds the font files: Noto's variable fonts fixed at Regular, as
// Gio draws a variable font's default instance (Thin for the CJK ones).
// Parsed, a font costs memory per glyph, so the CJK fonts keep only
// the common characters (JIS level 1 kanji with kana, all of Hangul,
// GB2312 hanzi, Big5 level 1) and the ones GoPlay's translations use,
// each Han character in the first font that has it, as fallback never
// reaches a later one.
//
//go:embed *.ttf
var FS embed.FS
