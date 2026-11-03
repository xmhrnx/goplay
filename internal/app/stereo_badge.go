package app

import (
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"
)

// stereoBadge marks 3D or 360° mode in the bar; clicking opens its menu.
func stereoBadge(a *App) elements.Element {
	label := "3D"
	if a.vrActive() {
		label = "360°"
	}
	return elements.Keyed("bar", "stereo-badge", func() *elements.Pressable {
		return elements.NewPressable(func(s elements.PressState) elements.Element {
			bg := colSurface3
			if s.Hovered {
				bg = colBorder
			}
			return elements.Box(styles.Styles{Background: bg, BorderRadius: 6, Padding: properties.SymmetricInset(8, 4), Margin: properties.SymmetricInset(4, 0)},
				elements.Text(label, styles.Styles{Color: colText, FontSize: 12, FontWeight: properties.FontWeightSemiBold}))
		}).Label("3D and 360° options").Focusable(false).OnClick(func(elements.ClickEvent) {
			a.ui.openMenuAt(a.ui.stereoMenu(), a.ui.lastPointer.Round())
		})
	})
}
