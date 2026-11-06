package app

import (
	"github.com/xmhrnx/giode/components/menu"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/player"
)

// Menus are rebuilt each time they open, so they show the current
// tracks, chapters and settings.

func newMenu() *menu.Menu {
	return menu.New(styles.Styles{
		Background:   colSurface,
		Color:        colText,
		BorderRadius: 10,
		Padding:      properties.UniformInset(6),
		BoxShadow:    properties.BoxShadow{Y: 10, Blur: 28, Color: "#000000d0"},
	}).
		ItemPadding(properties.SymmetricInset(14, 8)).
		HoverColor(colSurface3).
		SeparatorColor(colBorder).
		DisabledColor(colFaint)
}

// check appends an on/off row with a checkmark while on reports true,
// and radio one of a group of choices with a dot while it's the one
// chosen. Both are read every frame, so an open menu follows changes.
func check(m *menu.Menu, label string, on func() bool, fn func()) {
	m.CheckItem(label, on, fn)
}

func radio(m *menu.Menu, label string, on func() bool, fn func()) {
	m.RadioItem(label, on, fn)
}

// item appends a command by its ID: its title, greyed out when it can't
// run, checked when it's an active toggle.
func (u *ui) item(m *menu.Menu, id string) {
	act := u.a.actions.Get(id)
	if act == nil {
		return
	}
	switch {
	case !act.IsEnabled():
		m.DisabledItem(tr(act.Title))
	case act.Checked != nil:
		check(m, tr(act.Title), act.IsChecked, act.Run)
	default:
		m.Item(tr(act.Title), act.Run)
	}
}

func (u *ui) items(m *menu.Menu, ids ...string) {
	for _, id := range ids {
		if id == "-" {
			m.Separator()
			continue
		}
		u.item(m, id)
	}
}

// trackMenu lists one kind's tracks with "Off" for subtitles.
func (u *ui) trackMenu(m *menu.Menu, k player.TrackKind) {
	a := u.a
	st := &a.player.State
	tracks := st.TracksOf(k)
	selected := func(id int) func() bool {
		return func() bool {
			sel, ok := st.SelectedTrack(k)
			return ok == (id != 0) && (!ok || sel.ID == id)
		}
	}
	if k == player.Subtitle {
		radio(m, tr("Off"), selected(0), func() { a.pickTrack(k, 0) })
	}
	if len(tracks) == 0 && k != player.Subtitle {
		m.DisabledItem(tr("No tracks"))
		return
	}
	for _, t := range tracks {
		id := t.ID
		radio(m, ellipsize(t.Name(), 52), selected(id), func() { a.pickTrack(k, id) })
	}
}

func ellipsize(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
