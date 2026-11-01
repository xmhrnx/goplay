package app

import (
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"
)

// dropTarget shows that files dragged over the window can be dropped.
func dropTarget(playing bool) elements.Element {
	hint := tr("Drop to play")
	if playing {
		hint = tr("Drop to play, or to add subtitles")
	}
	return elements.Box(styles.Styles{Position: properties.PositionAbsolute, Top: 0, Left: 0, Right: 0, Bottom: 0,
		Padding: properties.UniformInset(12)},
		elements.Box(styles.Styles{Width: -1, Height: -1, Justify: properties.JustifyCenter, Align: properties.AlignCenter,
			Background: "#000000a6", BorderRadius: 16, BorderWidth: 2, BorderColor: colHighlight},
			elements.Box(styles.Styles{FlexDirection: properties.FlexDirectionColumn, Align: properties.AlignCenter, Gap: 12},
				ic("add", 40, colHighlight),
				elements.Text(hint, styles.Styles{Color: colText, FontSize: 16, FontWeight: properties.FontWeightMedium}))))
}
