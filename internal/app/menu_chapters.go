package app

import (
	"github.com/mehran9675/giode/components/menu"

	"github.com/mehran9675/goplay/internal/player"
)

// chapterMenu lists the chapters with their start times.
func (u *ui) chapterMenu() *menu.Menu {
	a := u.a
	st := &a.player.State
	m := newMenu()
	for i, c := range st.Chapters {
		i := i
		title := c.Title
		if title == "" {
			title = tr("Chapter {number}", "number", i+1)
		}
		radio(m, player.FormatTime(c.Time)+"   "+ellipsize(title, 40), func() bool { return i == st.Chapter }, func() { a.player.SetChapter(i) })
	}
	return m
}
