package app

import (
	"fmt"

	"github.com/xmhrnx/giode/components/menu"

	"github.com/mehran9675/goplay/internal/player"
	"github.com/mehran9675/goplay/internal/settings"
)

// audioMenu picks the audio track and its effects.
func (u *ui) audioMenu() *menu.Menu {
	a := u.a
	st := &a.player.State
	m := newMenu()
	u.trackMenu(m, player.Audio)
	m.Separator()
	u.items(m, "load_audio", "-", "audio_delay_down", "audio_delay_up")
	if st.AudioDelay != 0 {
		m.Item(tr("Reset Delay ({seconds} s)", "seconds", fmt.Sprintf("%+.1f", st.AudioDelay)), func() { a.player.SetAudioDelay(0); a.osd(tr("Audio delay 0")) })
	}
	m.Separator()
	eq := newMenu()
	d := settings.Lookup("equalizer")
	for _, o := range d.Options {
		v := o.Value
		radio(eq, tr(o.Label), func() bool { return a.settings.Get("equalizer") == v }, func() { a.setSetting("equalizer", v) })
	}
	m.SubMenu(tr("Equalizer"), eq)
	check(m, tr("Even Out Loudness"), func() bool { return a.settings.Bool("normalize") }, func() { a.toggleSetting("normalize", "Even out loudness") })
	check(m, tr("Night Mode"), func() bool { return a.settings.Bool("night_mode") }, func() { a.toggleSetting("night_mode", "Night mode") })
	m.Separator()
	m.Item(tr("Audio Settings…"), func() { u.dialogs.settings.show("Audio") })
	return m
}
