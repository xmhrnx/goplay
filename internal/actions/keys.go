package actions

import (
	"strings"

	"gioui.org/io/key"

	"github.com/xmhrnx/giode/shortcut"
)

// Shortcuts are written the way Giode's shortcut package reads them:
// modifiers then a key, joined by "+" ("Mod+O", "Shift+Left", "Space").
// "Mod" is Command on macOS and Ctrl elsewhere.

// combo is a parsed shortcut.
type combo struct {
	name key.Name
	mods key.Modifiers
}

func parse(spec string) (combo, bool) {
	name, mods, err := shortcut.Parse(spec)
	return combo{name, mods}, err == nil
}

// Same reports whether two shortcuts are the same keys, however they're
// spelled ("ctrl+o" and "Mod+O" off macOS).
func Same(a, b string) bool {
	ca, oka := parse(a)
	cb, okb := parse(b)
	return oka && okb && ca == cb
}

var keyNames = map[key.Name]string{
	key.NameLeftArrow:      "Left",
	key.NameRightArrow:     "Right",
	key.NameUpArrow:        "Up",
	key.NameDownArrow:      "Down",
	key.NameReturn:         "Enter",
	key.NameEnter:          "Enter",
	key.NameEscape:         "Escape",
	key.NameHome:           "Home",
	key.NameEnd:            "End",
	key.NameDeleteBackward: "Backspace",
	key.NameDeleteForward:  "Delete",
	key.NamePageUp:         "PageUp",
	key.NamePageDown:       "PageDown",
	key.NameTab:            "Tab",
	key.NameSpace:          "Space",
}

// FromEvent writes a key press as a shortcut.
func FromEvent(e key.Event) string {
	name, ok := keyNames[e.Name]
	if !ok {
		name = strings.ToUpper(string(e.Name))
	}
	m := e.Modifiers
	var parts []string
	for _, mod := range []struct {
		m    key.Modifiers
		name string
	}{{key.ModShortcut, "Mod"}, {key.ModCtrl, "Ctrl"}, {key.ModCommand, "Cmd"}, {key.ModSuper, "Super"},
		{key.ModAlt, "Alt"}, {key.ModShift, "Shift"}} {
		if m.Contain(mod.m) {
			parts = append(parts, mod.name)
			m &^= mod.m
		}
	}
	return strings.Join(append(parts, name), "+")
}

// MPVKey writes a key press in mpv's key syntax ("ctrl+LEFT", "A",
// "SPACE"), for forwarding unhandled keys to mpv's input.conf bindings.
func MPVKey(e key.Event) string {
	names := map[key.Name]string{
		key.NameLeftArrow: "LEFT", key.NameRightArrow: "RIGHT", key.NameUpArrow: "UP", key.NameDownArrow: "DOWN",
		key.NameReturn: "ENTER", key.NameEnter: "KP_ENTER", key.NameEscape: "ESC", key.NameHome: "HOME",
		key.NameEnd: "END", key.NameDeleteBackward: "BS", key.NameDeleteForward: "DEL", key.NamePageUp: "PGUP",
		key.NamePageDown: "PGDWN", key.NameTab: "TAB", key.NameSpace: "SPACE",
	}
	name, named := names[e.Name]
	m := e.Modifiers
	if !named {
		name = string(e.Name)
		if len(name) == 1 && name >= "A" && name <= "Z" {
			if m.Contain(key.ModShift) {
				m &^= key.ModShift // mpv writes Shift+a as "A"
			} else {
				name = strings.ToLower(name)
			}
		}
	}
	var prefix string
	if m.Contain(key.ModCtrl) {
		prefix += "ctrl+"
	}
	if m.Contain(key.ModAlt) {
		prefix += "alt+"
	}
	if m.Contain(key.ModCommand) || m.Contain(key.ModSuper) {
		prefix += "meta+"
	}
	if m.Contain(key.ModShift) {
		prefix += "shift+"
	}
	return prefix + name
}
