package app

// openAppMenu is the title bar's menu: the whole app, for anyone who
// doesn't know about right-clicking.
func (u *ui) openAppMenu() {
	m := newMenu()
	m.SubMenu(tr("Open"), u.openMenu())
	if !u.a.player.State.Idle {
		m.SubMenu(tr("Playback"), u.playbackMenu())
		m.SubMenu(tr("Video"), u.videoMenu())
		m.SubMenu(tr("Audio"), u.audioMenu())
		m.SubMenu(tr("Subtitles"), u.subtitleMenu())
		m.SubMenu(tr("3D & 360°"), u.stereoMenu())
	}
	m.Separator()
	u.items(m, "playlist", "cast", "-", "settings", "shortcuts", "palette", "-", "log", "config_folder", "about", "-", "quit")
	u.openMenuOver(m, u.appMenuAnchor)
}
