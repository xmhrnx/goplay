package app

import (
	"github.com/mehran9675/giode/components/button"
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"
)

// The pieces the dialogs are built from.

func dialogStyles(width int) styles.Styles {
	return styles.Styles{
		Background:   colSurface,
		BorderRadius: 14,
		Padding:      properties.UniformInset(22),
		Width:        properties.Width(width),
		BoxShadow:    properties.BoxShadow{Y: 18, Blur: 48, Color: "#000000e6"},
	}
}

func fieldStyles() styles.Styles {
	return styles.Styles{Background: colSurface2, Color: colText, BorderColor: colBorder, BorderRadius: 8, Padding: properties.SymmetricInset(12, 10)}
}

func primaryButton(label string, fn func()) *button.Button {
	return button.New(label, styles.Styles{Background: colHighlight, Color: colOnLight, FontSize: 14,
		FontWeight: properties.FontWeightSemiBold, Padding: properties.SymmetricInset(18, 9), BorderRadius: 8}).OnClick(fn)
}

func secondaryButton(label string, fn func()) *button.Button {
	return button.New(label, styles.Styles{Background: colSurface3, Color: colText, FontSize: 14,
		Padding: properties.SymmetricInset(18, 9), BorderRadius: 8}).OnClick(fn)
}

func heading(s string) elements.Element {
	return elements.Text(s, styles.Styles{Color: colText, FontSize: 18, FontWeight: properties.FontWeightSemiBold})
}

func note(s string) elements.Element {
	return elements.Text(s, styles.Styles{Color: colMuted, FontSize: 13})
}

func column(gap int, els ...elements.Element) elements.Element {
	return elements.Box(styles.Styles{Width: -1, FlexDirection: properties.FlexDirectionColumn, Gap: properties.Gap(gap)}, els...)
}

func buttonsRow(els ...elements.Element) elements.Element {
	all := append([]elements.Element{elements.Box(styles.Styles{FlexGrow: 1, Width: -1})}, els...)
	return elements.Row(styles.Styles{Width: -1, Gap: 8, Align: properties.AlignCenter, Wrap: properties.WrapNoWrap, Margin: properties.Inset{Top: 6}}, all...)
}

// keyChip shows a shortcut like a keycap.
func keyChip(text string, highlight bool) elements.Element {
	bg, fg := colSurface3, colText
	if highlight {
		bg, fg = colHighlight, colOnLight
	}
	return elements.Box(styles.Styles{Background: bg, BorderRadius: 5, Padding: properties.SymmetricInset(7, 3)},
		elements.Text(text, styles.Styles{Color: fg, FontSize: 12, FontWeight: properties.FontWeightMedium}))
}
