package app

import "github.com/xmhrnx/giode/components/menu"

// openMenu holds the ways to open media and what's open.
func (u *ui) openMenu() *menu.Menu {
	m := newMenu()
	u.items(m, "open", "open_folder", "open_disc", "open_url", "network_media", "paste", "-", "add_files", "load_subtitle", "load_audio")
	m.Separator()
	m.SubMenu(tr("Recent"), u.recentMenu())
	u.items(m, "-", "reveal", "copy_path", "close_media")
	return m
}
