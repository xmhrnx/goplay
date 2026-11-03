package app

import (
	"fmt"

	"github.com/xmhrnx/giode/components/menu"

	"github.com/mehran9675/goplay/internal/player"
)

// subtitleMenu picks the subtitles, a second line, and their timing.
func (u *ui) subtitleMenu() *menu.Menu {
	a := u.a
	st := &a.player.State
	m := newMenu()
	u.trackMenu(m, player.Subtitle)
	if subs := st.TracksOf(player.Subtitle); len(subs) > 1 {
		sec := newMenu()
		second := func(id int) func() bool {
			return func() bool {
				cur, ok := st.SecondarySubtitle()
				return ok == (id != 0) && (!ok || cur.ID == id)
			}
		}
		radio(sec, tr("Off"), second(0), func() { a.player.SelectSecondarySubtitle(0) })
		for _, t := range subs {
			id := t.ID
			radio(sec, ellipsize(t.Name(), 52), second(id), func() { a.player.SelectSecondarySubtitle(id) })
		}
		m.Separator()
		m.SubMenu(tr("Second Subtitle"), sec)
	}
	m.Separator()
	u.items(m, "load_subtitle", "-", "sub_delay_down", "sub_delay_up", "sub_smaller", "sub_bigger")
	if st.SubDelay != 0 {
		m.Item(tr("Reset Delay ({seconds} s)", "seconds", fmt.Sprintf("%+.1f", st.SubDelay)), func() { a.player.SetSubDelay(0); a.osd(tr("Subtitle delay 0")) })
	}
	m.Separator()
	m.Item(tr("Subtitle Settings…"), func() { u.dialogs.settings.show("Subtitles") })
	return m
}
