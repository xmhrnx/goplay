package app

import (
	"github.com/mehran9675/giode/components/menu"

	"github.com/mehran9675/goplay/internal/library"
)

// editionMenu lists a disc's titles, or a file's editions.
func (u *ui) editionMenu() *menu.Menu {
	a := u.a
	st := &a.player.State
	disc := library.IsDiscURL(st.Path)
	m := newMenu()
	for i, e := range st.Editions {
		id := e.ID
		title := e.Title
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
