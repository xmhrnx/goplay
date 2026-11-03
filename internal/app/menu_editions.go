package app

import (
	"time"

	"github.com/mehran9675/giode/components/menu"

	"github.com/mehran9675/goplay/internal/library"
	"github.com/mehran9675/goplay/internal/player"
)

// editionMenu lists a disc's titles, or a file's editions.
func (u *ui) editionMenu() *menu.Menu {
	a := u.a
	st := &a.player.State
	disc := library.IsDiscURL(st.Path)
	// A Blu-ray lists its titles with their lengths: the short ones
	// (menus, logos, warnings) are left out, and the longest is the movie.
	longest, most := -1, time.Duration(0)
	for _, e := range st.Editions {
		if t, ok := library.ParseDiscTitle(e.Title); ok && t.Length > most {
			longest, most = e.ID, t.Length
		}
	}
	m := newMenu()
	for i, e := range st.Editions {
		id := e.ID
		title := e.Title
		if t, ok := library.ParseDiscTitle(e.Title); ok {
			if t.Length < library.MinDiscTitle && id != st.Edition {
				continue
			}
			title = tr("Title {number}", "number", t.Number) + " · " + player.FormatTime(t.Length.Seconds())
			if id == longest {
				title = tr("Main movie") + " · " + player.FormatTime(t.Length.Seconds())
			}
		}
		switch {
		case title != "":
		case disc:
			title = tr("Title {number}", "number", i+1)
		default:
			title = tr("Edition {number}", "number", i+1)
		}
		radio(m, ellipsize(title, 48), func() bool { return id == st.Edition }, func() { a.player.SetEdition(id) })
	}
	return m
}
