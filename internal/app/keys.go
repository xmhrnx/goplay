package app

import (
	"gioui.org/io/key"

	"github.com/mehran9675/goplay/internal/actions"
	"github.com/mehran9675/goplay/internal/settings"
)

// onKey takes the keys nothing in the UI took (text fields, dialogs and
// menus come first): shortcuts, and otherwise mpv's own bindings.
func (a *App) onKey(e key.Event) bool {
	if e.State != key.Press || !a.handleKey(e) {
		return false
	}
	a.dirty = true
	return true
}

// handleKey runs the shortcut for a key press. It reports whether the
// key was used.
func (a *App) handleKey(e key.Event) bool {
	if a.ui.capturingKey() {
		a.ui.captureKey(e)
		return true
	}
	if a.ui.modalOpen() && e.Name != key.NameEscape {
		return false
	}
	if e.Name == key.NameEscape {
		return a.escape()
	}
	if act := a.actions.Lookup(e); act != nil {
		a.actions.Run(act.ID)
		return true
	}
	if a.settings.Bool(settings.KeyPassKeys) && !a.player.State.Idle && isPlainKey(e) {
		a.player.Keypress(actions.MPVKey(e))
		return true
	}
	return false
}

// isPlainKey reports whether a key is worth forwarding to mpv: not a
// lone modifier, Tab or Enter (which move focus and activate UI).
func isPlainKey(e key.Event) bool {
	switch e.Name {
	case key.NameTab, key.NameReturn, key.NameEnter, key.NameCtrl, key.NameShift, key.NameAlt, key.NameSuper, key.NameCommand:
		return false
	}
	return true
}

// escape closes the innermost thing open: a menu, dialog or panel, then
// fullscreen.
func (a *App) escape() bool {
	switch {
	case a.ui.closeTopmost():
		return true
	case a.g.Fullscreen():
		a.setFullscreen(false)
		return true
	case a.ui.pip:
		a.togglePiP()
		return true
	}
	return false
}

// afterPlayerChange reacts to state changes: keeping the display awake
// while playing, the window title, remembering progress.
func (a *App) afterPlayerChange() {
	st := &a.player.State
	playing := !st.Idle && !st.Paused && !a.casting()
	a.g.KeepScreenOn(playing && st.HasVideo())
	title := "GoPlay"
	if !st.Idle && st.Title != "" {
		title = st.Title + " — GoPlay"
	}
	if title != a.title {
		a.title = title
		a.g.SetTitle(title)
	}
	if !st.Idle && st.Path != "" {
		a.lastPath, a.lastTitle = st.Path, st.Title
	}
}
