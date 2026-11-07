package app

import (
	"gioui.org/io/key"

	"github.com/xmhrnx/giode/shortcut"

	"github.com/mehran9675/goplay/internal/actions"
)

func (u *ui) capturingKey() bool {
	return u.dialogs.settings.open && u.dialogs.settings.keys.capture != ""
}

// captureKey records a key press as the shortcut being edited.
func (u *ui) captureKey(ev key.Event) {
	e := u.dialogs.settings.keys
	a := u.a
	switch ev.Name {
	case key.NameCtrl, key.NameShift, key.NameAlt, key.NameSuper, key.NameCommand:
		return // wait for the key itself
	}
	id, idx := e.capture, e.captureIdx
	e.capture = ""
	keys := append([]string(nil), a.actions.Keys(id)...)
	if (ev.Name == key.NameDeleteBackward || ev.Name == key.NameDeleteForward) && ev.Modifiers == 0 {
		if idx >= 0 && idx < len(keys) {
			keys = append(keys[:idx], keys[idx+1:]...)
			a.actions.SetCustom(id, keys)
			a.settings.SetKeys(id, keys)
		}
		return
	}
	sc := actions.FromEvent(ev)
	if owner := a.actions.Owner(sc); owner != nil && owner.ID != id {
		rest := []string{} // empty, not nil: nil restores the defaults
		for _, k := range a.actions.Keys(owner.ID) {
			if !actions.Same(k, sc) {
				rest = append(rest, k)
			}
		}
		a.actions.SetCustom(owner.ID, rest)
		a.settings.SetKeys(owner.ID, rest)
		a.toast(tr("{shortcut} moved from “{command}”", "shortcut", shortcut.Format(sc), "command", tr(owner.Title)))
	}
	if idx >= 0 && idx < len(keys) {
		keys[idx] = sc
	} else {
		keys = append(keys, sc)
	}
	a.actions.SetCustom(id, keys)
	a.settings.SetKeys(id, keys)
}
