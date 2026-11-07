package app

import (
	"math"

	"github.com/xmhrnx/giode/components/iconbutton"
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/cast"
)

// ghostButton is a round, transparent icon button that lights up on
// hover.
func ghostButton(icon string, size int, label string, onClick func()) *iconbutton.Button {
	pad := (buttonSize - size) / 2
	return iconbutton.New(ic(icon, size, colText), styles.Styles{
		Color:        colText,
		Width:        buttonSize,
		Height:       buttonSize,
		Padding:      properties.UniformInset(pad),
		BorderRadius: buttonSize / 2,
	}).Variant(iconbutton.VariantGhost).Label(label).OnClick(onClick)
}

// refresh updates icons, labels and the volume slider from the state.
func (b *controlBar) refresh() {
	a := b.u.a
	st := &a.player.State
	playing := !st.Idle && !st.Paused
	icon := "play"
	switch {
	case a.casting():
		if s := a.cast.status.State; s == cast.StatePlaying || s == cast.StateBuffering || s == cast.StateLoading {
			icon = "pause"
		}
	case st.EOF && st.Paused:
		icon = "replay"
	case playing:
		icon = "pause"
	}
	b.play.Icon(ic(icon, playIconSize, colOnLight))

	vol, max := st.Volume, st.VolumeMax
	if max <= 0 {
		max = 100
	}
	level := vol / max
	if a.casting() {
		level = math.Max(0, a.cast.status.Volume)
		vol = level * 100
		max = 100
	}
	mic := "volume-high"
	switch {
	case st.Muted || vol <= 0:
		mic = "volume-off"
	case vol < 50:
		mic = "volume-low"
	}
	b.mute.Icon(ic(mic, iconSize, colText))
	// The audio effects light up while on.
	effect := func(btn *iconbutton.Button, icon, setting string) {
		c := colFaint
		if a.settings.Bool(setting) {
			c = colHighlight
		}
		btn.Icon(ic(icon, iconSize, c))
	}
	effect(b.normalize, "normalize", "normalize")
	effect(b.night, "night", "night_mode")
	if !b.vol.Dragging() {
		b.vol.SetValue(float32(level))
	}

	full := "fullscreen"
	if a.g.Fullscreen() {
		full = "fullscreen-exit"
	}
	b.full.Icon(ic(full, iconSize, colText))
	castIcon := "cast"
	if a.casting() {
		castIcon = "cast-on"
	}
	b.cast.Icon(ic(castIcon, iconSize, colText))
	b.speed.Label(speedLabel(st.Speed))
	for btn, t := range b.tips {
		t.SetContent(b.tipText[btn]())
	}
}

func spacer(w int) elements.Element { return elements.Box(styles.Styles{Width: properties.Width(w)}) }

// timeText is a time readout with tabular digits so it doesn't jitter.
func timeText(s, color string) elements.Element {
	return elements.Text(s, styles.Styles{Color: color, FontSize: 13, FontWeight: properties.FontWeightMedium})
}
