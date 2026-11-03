package app

import (
	"github.com/xmhrnx/giode/components/progress"
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"
)

// osdView is the transient message at the top of the picture.
func osdView(u *ui) elements.Element {
	top := 20
	if u.overlay() {
		top += u.titleHeightDp()
	}
	children := []elements.Element{elements.Text(u.osdText, styles.Styles{Color: colText, FontSize: 15, FontWeight: properties.FontWeightMedium})}
	if u.osdLevel >= 0 {
		children = append(children, progress.Progress(float32(u.osdLevel), styles.Styles{
			Width: 160, Height: 4, Color: colHighlight, Background: colTrack, BorderRadius: 2,
		}))
	}
	return elements.Box(styles.Styles{Position: properties.PositionAbsolute, Top: properties.Top(top), Left: 0, Right: 0,
		Width: -1, Justify: properties.JustifyCenter},
		elements.Box(styles.Styles{
			Background: "#121214d9", BorderRadius: 12, Padding: properties.SymmetricInset(16, 10),
			FlexDirection: properties.FlexDirectionColumn, Align: properties.AlignCenter, Gap: 8,
			BoxShadow: properties.BoxShadow{Y: 4, Blur: 16, Color: "#00000080"},
		}, children...))
}
