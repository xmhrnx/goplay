package app

import (
	"github.com/mehran9675/giode/components/button"
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/cast"
)

// castScreen replaces the picture while the video plays on a TV.
func castScreen(a *App) elements.Element {
	c := &a.cast
	state := map[cast.State]string{
		cast.StateLoading: tr("Loading on"), cast.StatePlaying: tr("Playing on"), cast.StatePaused: tr("Paused on"),
		cast.StateBuffering: tr("Buffering on"), cast.StateIdle: tr("Connected to"), cast.StateStopped: tr("Finished on"),
	}[c.status.State]
	if state == "" {
		state = tr("Casting to")
	}
	stop := elements.Keyed("cast", "stop", func() *button.Button {
		return button.New(tr("Stop Casting"), styles.Styles{
			Background: colSurface3, Color: colText, FontSize: 14, Padding: properties.SymmetricInset(18, 10), BorderRadius: 20,
		}).OnClick(a.stopCasting)
	})
	return centered(elements.Box(styles.Styles{FlexDirection: properties.FlexDirectionColumn, Align: properties.AlignCenter, Gap: 12},
		ic("cast-on", 64, colText),
		elements.Text(state, styles.Styles{Color: colMuted, FontSize: 14}),
		elements.Text(c.device, styles.Styles{Color: colText, FontSize: 22, FontWeight: properties.FontWeightSemiBold}),
		elements.Text(displayTitle(&a.player.State), styles.Styles{Color: colMuted, FontSize: 13, MaxLines: 1}),
		elements.Box(styles.Styles{Margin: properties.Inset{Top: 8}}, stop),
	))
}
