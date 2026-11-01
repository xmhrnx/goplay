package app

import (
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/player"
)

// audioOnly is shown for files without video: a large note and the
// title.
func audioOnly(st *player.State) elements.Element {
	return centered(elements.Box(styles.Styles{FlexDirection: properties.FlexDirectionColumn, Align: properties.AlignCenter, Gap: 14},
		elements.Box(styles.Styles{Background: colSurface2, BorderRadius: 24, Padding: properties.UniformInset(28)}, ic("music-large", 72, colMuted)),
		elements.Text(displayTitle(st), styles.Styles{Color: colText, FontSize: 18, FontWeight: properties.FontWeightSemiBold, MaxLines: 2, TextAlign: properties.TextAlignCenter}),
	))
}
