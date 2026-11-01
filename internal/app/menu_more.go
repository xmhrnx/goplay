package app

import "github.com/mehran9675/giode/components/menu"

// moreMenu is the bar's overflow: the less frequent controls.
func (u *ui) moreMenu() *menu.Menu {
	m := newMenu()
	m.SubMenu(tr("Video"), u.videoMenu())
	m.SubMenu(tr("3D & 360°"), u.stereoMenu())
	m.SubMenu(tr("Playback"), u.playbackMenu())
	m.Separator()
	u.items(m, "screenshot", "info", "pip", "always_on_top", "-", "settings")
	return m
}
