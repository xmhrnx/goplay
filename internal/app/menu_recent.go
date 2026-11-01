package app

import "github.com/mehran9675/giode/components/menu"

// recentMenu lists the recently played files.
func (u *ui) recentMenu() *menu.Menu {
	a := u.a
	m := newMenu()
	items := a.history.Items()
	if len(items) == 0 {
		m.DisabledItem(tr("Nothing yet"))
		return m
	}
	for i, r := range items {
		if i == 12 {
			break
		}
		path := r.Path
		m.Item(ellipsize(r.Name(), 48), func() { a.openItems([]string{path}, false) })
	}
	m.Separator()
	m.Item(tr("Clear Recent"), func() { a.history.Clear(); a.history.Save() })
	return m
}
