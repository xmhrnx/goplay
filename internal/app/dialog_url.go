package app

import (
	"strings"

	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/elements"

	"github.com/mehran9675/goplay/internal/library"
)

// urlDialog asks for a link to play.
type urlDialog struct {
	modal
	u     *ui
	input *input.Input
}

func newURLDialog(u *ui) *urlDialog {
	a := u.a
	d := &urlDialog{u: u}
	d.input = input.New("https://…", fieldStyles()).Keyboard(input.KeyboardURL).
		Submit(func(s string) { d.open = false; a.openURL(s) })
	openBtn := primaryButton(tr("Open"), func() { d.open = false; a.openURL(d.input.Text()) })
	cancel := secondaryButton(tr("Cancel"), func() { d.open = false })
	d.dlg = dialog.New(&d.open, func() elements.Element {
		return column(12,
			heading(tr("Open URL")),
			note(tr("A video or stream link, or a page from YouTube, Vimeo, Twitch and thousands of other sites.")),
			d.input,
			buttonsRow(cancel, openBtn))
	}, dialogStyles(520))
	return d
}

func (d *urlDialog) show() {
	d.input.SetText("")
	d.open = true
	d.input.Focus()
	// A link on the clipboard fills the field.
	d.u.a.g.ReadClipboard(func(text string) {
		text = strings.TrimSpace(text)
		if d.open && d.input.Text() == "" && library.IsURL(text) {
			d.input.SetText(text)
			d.input.SelectAll()
		}
	})
}
