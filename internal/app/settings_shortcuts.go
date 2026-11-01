package app

import (
	"strconv"

	"gioui.org/io/key"

	"github.com/mehran9675/giode/components/button"
	"github.com/mehran9675/giode/components/scroll"
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/actions"
)

// shortcutEditor is the settings dialog's Shortcuts page: every command
// with its shortcuts, changed by clicking one and pressing new keys.
type shortcutEditor struct {
	u        *ui
	scroll   *scroll.Scroll
	resetAll *button.Button

	// capture is the action whose shortcut is being recorded, and which
	// of its shortcuts (-1 adds one).
	capture    string
	captureIdx int
}

func newShortcutEditor(u *ui) *shortcutEditor {
	a := u.a
	e := &shortcutEditor{u: u}
	e.scroll = scroll.New(func() elements.Element { return clearOfBar(e.content()) }, scrollStyles())
	e.resetAll = secondaryButton(tr("Reset All Shortcuts"), func() {
		a.settings.ResetKeys()
		for _, act := range a.actions.All() {
			a.actions.SetCustom(act.ID, nil)
		}
		a.toast(tr("Shortcuts reset"))
	})
	return e
}

// content lists every command with its shortcuts.
func (e *shortcutEditor) content() elements.Element {
	a := e.u.a
	rows := []elements.Element{note(tr("Click a shortcut to change it, then press the new keys. Backspace removes it; Escape cancels."))}
	for _, cat := range a.actions.Categories() {
		rows = append(rows, elements.Box(styles.Styles{Padding: properties.Inset{Top: 14, Bottom: 4}},
			elements.Text(tr(cat), styles.Styles{Color: colMuted, FontSize: 12, FontWeight: properties.FontWeightSemiBold})))
		for _, act := range a.actions.All() {
			if act.Category != cat || act.Hidden {
				continue
			}
			rows = append(rows, e.row(act))
		}
	}
	return elements.Box(styles.Styles{Width: -1, FlexDirection: properties.FlexDirectionColumn, Gap: 2, Padding: properties.Inset{Right: 12, Bottom: 12}}, rows...)
}

func (e *shortcutEditor) row(act *actions.Action) elements.Element {
	a := e.u.a
	keys := a.actions.Keys(act.ID)
	chips := []elements.Element{}
	for i, k := range keys {
		chips = append(chips, e.chip(act.ID, i, actions.Display(k)))
	}
	if e.capture == act.ID && e.captureIdx < 0 {
		chips = append(chips, keyChip(tr("Press keys…"), true))
	} else if len(keys) < 3 {
		chips = append(chips, e.chip(act.ID, -1, "+"))
	}
	if a.actions.Customized(act.ID) {
		chips = append(chips, elements.Keyed("shortcut-reset", act.ID, func() *elements.Pressable {
			return elements.NewPressable(func(s elements.PressState) elements.Element {
				c := colFaint
				if s.Hovered {
					c = colText
				}
				return elements.Text(tr("Reset"), styles.Styles{Color: c, FontSize: 12})
			}).Label(tr("Reset {command}", "command", tr(act.Title))).OnClick(func(elements.ClickEvent) {
				a.actions.SetCustom(act.ID, nil)
				a.settings.SetKeys(act.ID, nil)
			})
		}))
	}
	return elements.Row(styles.Styles{Width: -1, Gap: 6, Align: properties.AlignCenter, Wrap: properties.WrapNoWrap, Padding: properties.SymmetricInset(0, 5)},
		append([]elements.Element{elements.Box(styles.Styles{FlexGrow: 1, Width: -1}, elements.Text(tr(act.Title), styles.Styles{Color: colText, FontSize: 13}))}, chips...)...)
}

// chip is a clickable shortcut; i is its index, or -1 for "add".
func (e *shortcutEditor) chip(id string, i int, text string) elements.Element {
	capturing := e.capture == id && e.captureIdx == i
	if capturing {
		text = tr("Press keys…")
	}
	return elements.Keyed("shortcut-chip", id+"#"+strconv.Itoa(i), func() *elements.Pressable {
		return elements.NewPressable(nil)
	}).Content(func(s elements.PressState) elements.Element {
		return keyChip(text, capturing || s.Hovered)
	}).Label(tr("Change shortcut")).Focusable(false).OnClick(func(elements.ClickEvent) {
		e.capture, e.captureIdx = id, i
	})
}

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
	case key.NameEscape:
		e.capture = ""
		return
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
		var rest []string
		for _, k := range a.actions.Keys(owner.ID) {
			if actions.Normalize(k) != sc {
				rest = append(rest, k)
			}
		}
		a.actions.SetCustom(owner.ID, rest)
		a.settings.SetKeys(owner.ID, rest)
		a.toast(tr("{shortcut} moved from “{command}”", "shortcut", actions.Display(sc), "command", tr(owner.Title)))
	}
	if idx >= 0 && idx < len(keys) {
		keys[idx] = sc
	} else {
		keys = append(keys, sc)
	}
	a.actions.SetCustom(id, keys)
	a.settings.SetKeys(id, keys)
}
