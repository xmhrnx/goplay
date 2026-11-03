package app

import (
	"math"

	"github.com/xmhrnx/giode/components/menu"
)

// speedMenu lists the playback speeds.
func (u *ui) speedMenu() *menu.Menu {
	a := u.a
	st := &a.player.State
	m := newMenu()
	for _, s := range speedSteps {
		label := speedLabel(s)
		if s == 1 {
			label = tr("Normal")
		}
		radio(m, label, func() bool { return math.Abs(s-st.Speed) < 1e-6 }, func() { a.setSpeed(s) })
	}
	return m
}
