package app

import (
	"github.com/mehran9675/giode/components/dialog"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/components/virtuallist"
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/actions"
)

// paletteDialog is the command palette: every command, found by typing.
type paletteDialog struct {
	modal
	u       *ui
	input   *input.Input
	list    *virtuallist.List
	matches []*actions.Action
}

func newPaletteDialog(u *ui) *paletteDialog {
	a := u.a
	d := &paletteDialog{u: u}
	d.input = input.New(tr("Type a command…"), fieldStyles()).
		Submit(func(string) { d.run(0) })
	d.list = virtuallist.New(func() int { return len(d.matches) }, 40, gapped(d.row), listStyles()).
		OnActivate(d.run).ClickFocusesRow(false)
	d.dlg = dialog.New(&d.open, func() elements.Element {
		d.matches = a.actions.Search(d.input.Text())
		h := min(len(d.matches), 9) * 40
		if h == 0 {
			return column(10, d.input, note(tr("No matching command.")))
		}
		return column(10, d.input, elements.Box(styles.Styles{Width: -1, Height: properties.Height(h)}, d.list))
	}, dialogStyles(560))
	return d
}

func (d *paletteDialog) show() {
	d.input.SetText("")
	d.open = true
	d.input.Focus()
}

func (d *paletteDialog) run(i int) {
	if i < 0 || i >= len(d.matches) {
		return
	}
	act := d.matches[i]
	d.open = false
	if act.IsEnabled() {
		elements.AfterLayout(act.Run)
	}
}

func (d *paletteDialog) row(i int) elements.Element {
	if i >= len(d.matches) {
		return elements.Box(styles.Styles{})
	}
	act := d.matches[i]
	a := d.u.a
	return elements.Keyed("palette", act.ID, func() *elements.Pressable {
		return elements.NewPressable(nil)
	}).Content(func(s elements.PressState) elements.Element {
		bg := ""
		if s.Hovered || i == 0 {
			bg = colSurface2
		}
		titleColor := colText
		if !act.IsEnabled() {
			titleColor = colFaint
		}
		children := []elements.Element{
			elements.Box(styles.Styles{FlexGrow: 1, Width: -1}, elements.Text(tr(act.Title), styles.Styles{Color: titleColor, FontSize: 14})),
			elements.Text(tr(act.Category), styles.Styles{Color: colFaint, FontSize: 12}),
		}
		if ks := a.actions.Keys(act.ID); len(ks) > 0 {
			children = append(children, keyChip(actions.Display(ks[0]), false))
		}
		return elements.Row(styles.Styles{Width: -1, Height: -1, Gap: 10, Align: properties.AlignCenter, Padding: properties.SymmetricInset(10, 0),
			Background: bg, BorderRadius: 8, Wrap: properties.WrapNoWrap}, children...)
	}).Label(tr(act.Title)).OnClick(func(elements.ClickEvent) { d.run(i) })
}
