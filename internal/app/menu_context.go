package app

import (
	"github.com/mehran9675/giode/components/menu"

	"github.com/mehran9675/goplay/internal/library"
)

// contextMenu is the right-click menu: everything, grouped.
func (u *ui) contextMenu() *menu.Menu {
	a := u.a
	st := &a.player.State
	m := newMenu()
	if st.Idle {
		u.items(m, "open", "open_folder", "open_disc", "open_url", "network_media", "paste")
		m.SubMenu(tr("Open Recent"), u.recentMenu())
		m.Separator()
		u.items(m, "settings", "shortcuts", "palette", "-", "about", "quit")
		return m
	}
	u.items(m, "play_pause", "stop")
	m.Separator()
	if library.IsDiscURL(st.Path) && len(st.Editions) > 1 {
		// A disc's extras are other titles: shown first, not only
		// under Playback.
		m.SubMenu(tr("Titles"), u.editionMenu())
	}
	m.SubMenu(tr("Playback"), u.playbackMenu())
	m.SubMenu(tr("Video"), u.videoMenu())
	m.SubMenu(tr("Audio"), u.audioMenu())
	m.SubMenu(tr("Subtitles"), u.subtitleMenu())
	m.SubMenu(tr("3D & 360°"), u.stereoMenu())
	m.Separator()
	m.SubMenu(tr("Open"), u.openMenu())
	u.items(m, "playlist", "cast", "-", "fullscreen", "pip", "always_on_top", "-", "settings", "palette")
	return m
}
