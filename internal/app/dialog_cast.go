package app

import (
	"gioui.org/layout"

	"github.com/mehran9675/giode/components/button"
	"github.com/mehran9675/giode/components/dialog"
	"github.com/mehran9675/giode/components/spinner"
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/cast"
)

// castDialog lists the TVs and speakers to cast to.
type castDialog struct {
	modal
	u                      *ui
	rescan, stop, closeBtn *button.Button
}

func newCastDialog(u *ui) *castDialog {
	a := u.a
	d := &castDialog{u: u}
	d.rescan = secondaryButton(tr("Search Again"), a.startCastScan)
	d.stop = secondaryButton(tr("Stop Casting"), func() { a.stopCasting(); d.open = false })
	d.closeBtn = secondaryButton(tr("Close"), func() { d.open = false })
	d.dlg = dialog.New(&d.open, d.content, dialogStyles(460))
	return d
}

func (d *castDialog) show() {
	d.open = true
	if !d.u.a.casting() {
		d.u.a.startCastScan()
	}
}

func (d *castDialog) layout(gtx layout.Context) {
	d.dlg.Layout(gtx)
	if !d.open && d.u.a.cast.scanning {
		d.u.a.stopCastScan()
	}
}

func (d *castDialog) content() elements.Element {
	a := d.u.a
	c := &a.cast
	rows := []elements.Element{heading(tr("Cast to a device"))}
	if a.casting() {
		rows = append(rows, note(tr("Playing on {device}.", "device", c.device)), buttonsRow(d.stop, d.closeBtn))
		return column(12, rows...)
	}
	rows = append(rows, note(tr("Chromecast, Google TV and Android TV, and smart TVs, consoles and receivers that support DLNA.")))
	if c.connecting != "" {
		rows = append(rows, elements.Row(styles.Styles{Gap: 10, Align: properties.AlignCenter},
			spinner.Spinner(styles.Styles{Width: 18, Height: 18, Color: colText}), note(tr("Connecting to {device}…", "device", c.connecting))))
	}
	for _, dev := range c.devices {
		rows = append(rows, d.deviceRow(dev))
	}
	if c.scanning {
		rows = append(rows, elements.Row(styles.Styles{Gap: 10, Align: properties.AlignCenter, Padding: properties.SymmetricInset(4, 6)},
			spinner.Spinner(styles.Styles{Width: 18, Height: 18, Color: colMuted}), note(tr("Looking for devices on your network…"))))
	} else if len(c.devices) == 0 {
		rows = append(rows, note(tr("No devices found. Make sure they're on and on the same network as this computer.")))
	}
	rows = append(rows, buttonsRow(d.rescan, d.closeBtn))
	return column(10, rows...)
}

func (d *castDialog) deviceRow(dev cast.Device) elements.Element {
	a := d.u.a
	return elements.Keyed("cast-device", dev.ID(), func() *elements.Pressable {
		return elements.NewPressable(func(s elements.PressState) elements.Element {
			bg := colSurface2
			if s.Hovered {
				bg = colSurface3
			}
			icon := "tv"
			if dev.Kind() == "Chromecast" {
				icon = "cast"
			}
			sub := dev.Kind()
			if dev.Model() != "" {
				sub += " · " + dev.Model()
			}
			return elements.Row(styles.Styles{Width: -1, Gap: 12, Align: properties.AlignCenter, Padding: properties.SymmetricInset(12, 10),
				Background: bg, BorderRadius: 10, Wrap: properties.WrapNoWrap},
				ic(icon, 22, colText),
				elements.Box(styles.Styles{FlexGrow: 1, Width: -1, FlexDirection: properties.FlexDirectionColumn, Gap: 2},
					elements.Text(dev.Name(), styles.Styles{Color: colText, FontSize: 14, FontWeight: properties.FontWeightMedium}),
					elements.Text(sub, styles.Styles{Color: colFaint, FontSize: 12})),
			)
		}).Label(dev.Name()).OnClick(func(elements.ClickEvent) {
			d.open = false
			a.castTo(dev)
		})
	})
}
