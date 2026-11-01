package app

import (
	"github.com/mehran9675/giode/components/menu"

	"github.com/mehran9675/goplay/internal/library"
)

// playbackMenu has speed, titles, chapters and moving through the
// file and the playlist.
func (u *ui) playbackMenu() *menu.Menu {
	a := u.a
	st := &a.player.State
	m := newMenu()
	m.SubMenu(tr("Speed"), u.speedMenu())
	if len(st.Editions) > 1 {
		label := tr("Editions")
		if library.IsDiscURL(st.Path) {
			label = tr("Titles")
		}
		m.SubMenu(label, u.editionMenu())
	}
	if len(st.Chapters) > 0 {
		m.SubMenu(tr("Chapters"), u.chapterMenu())
	}
	u.items(m, "-", "previous", "next", "chapter_previous", "chapter_next", "-",
		"jump_to", "restart", "frame_back", "frame_forward", "-",
		"ab_loop", "loop_file", "loop_playlist", "shuffle")
	return m
}
