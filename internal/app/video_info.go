package app

import (
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"
)

// infoPanel is the media information panel.
func infoPanel(u *ui) elements.Element {
	top := 16
	if u.overlay() {
		top += u.titleHeightDp()
	}
	rows := []elements.Element{elements.Text(tr("Media information"), styles.Styles{Color: colText, FontSize: 13, FontWeight: properties.FontWeightSemiBold})}
	for _, r := range u.infoRows {
		rows = append(rows, elements.Row(styles.Styles{Gap: 12, Wrap: properties.WrapNoWrap},
			elements.Box(styles.Styles{Width: 120}, elements.Text(tr(r[0]), styles.Styles{Color: colMuted, FontSize: 12})),
			elements.Box(styles.Styles{MaxWidth: 300}, elements.Text(r[1], styles.Styles{Color: colText, FontSize: 12, MaxLines: 2})),
		))
	}
	return elements.Box(styles.Styles{
		Position: properties.PositionAbsolute, Top: properties.Top(top), Left: 16,
		Background: "#121214e6", BorderRadius: 12, Padding: properties.UniformInset(14),
		FlexDirection: properties.FlexDirectionColumn, Gap: 6,
	}, rows...)
}
