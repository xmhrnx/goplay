package app

import (
	"gioui.org/layout"

	"github.com/mehran9675/giode/components/dialog"
)

// dialogs are the modal windows, each in its own dialog_*.go file.
type dialogs struct {
	url      *urlDialog
	jump     *jumpDialog
	cast     *castDialog
	net      *netMedia
	palette  *paletteDialog
	log      *logDialog
	about    *aboutDialog
	settings *prefsView
}

// modal is what every dialog has: its open flag and the Giode dialog
// that watches it.
type modal struct {
	open bool
	dlg  *dialog.Dialog
}

func (m *modal) layout(gtx layout.Context) { m.dlg.Layout(gtx) }
func (m *modal) isOpen() bool              { return m.open }
func (m *modal) close()                    { m.open = false }

type window interface {
	layout(gtx layout.Context)
	isOpen() bool
	close()
}

func newDialogs(u *ui) *dialogs {
	return &dialogs{
		url:      newURLDialog(u),
		jump:     newJumpDialog(u),
		cast:     newCastDialog(u),
		net:      newNetMedia(u),
		palette:  newPaletteDialog(u),
		log:      newLogDialog(u),
		about:    newAboutDialog(u),
		settings: newPrefsView(u),
	}
}

// all lists the dialogs, the one closed first by Escape first.
func (d *dialogs) all() []window {
	return []window{d.settings, d.palette, d.url, d.jump, d.cast, d.net, d.log, d.about}
}

func (d *dialogs) layout(gtx layout.Context) {
	for _, w := range d.all() {
		w.layout(gtx)
	}
}

func (d *dialogs) anyOpen() bool {
	for _, w := range d.all() {
		if w.isOpen() {
			return true
		}
	}
	return false
}

// closeTop closes the open dialog; false if none was.
func (d *dialogs) closeTop() bool {
	for _, w := range d.all() {
		if w.isOpen() {
			w.close()
			return true
		}
	}
	return false
}
