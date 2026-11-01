package app

import (
	"strings"

	"github.com/mehran9675/giode/components/dialog"
	"github.com/mehran9675/giode/components/virtuallist"
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"
)

// logDialog shows mpv's recent log lines.
type logDialog struct {
	modal
	list  *virtuallist.List
	lines []string
}

func newLogDialog(u *ui) *logDialog {
	a := u.a
	d := &logDialog{}
	copyLog := secondaryButton(tr("Copy"), func() { a.g.SetClipboard(strings.Join(d.lines, "\n")); a.toast(tr("Log copied")) })
	closeLog := secondaryButton(tr("Close"), func() { d.open = false })
	d.list = virtuallist.New(func() int { return len(d.lines) }, 20, gapped(func(i int) elements.Element {
		if i >= len(d.lines) {
			return elements.Box(styles.Styles{})
		}
		return elements.Text(d.lines[i], styles.Styles{Color: colMuted, FontSize: 11, MaxLines: 1, TextOverflow: properties.TextOverflowEllipsis})
	}), listStyles())
	d.dlg = dialog.New(&d.open, func() elements.Element {
		d.lines = a.logLines()
		body := elements.Element(elements.Box(styles.Styles{Width: -1, Height: 420}, d.list))
		if len(d.lines) == 0 {
			body = note(tr("Nothing logged yet. Raise “mpv log detail” in Settings › Advanced for more."))
		}
		return column(12, heading(tr("mpv Log")), body, buttonsRow(copyLog, closeLog))
	}, dialogStyles(760))
	return d
}

func (d *logDialog) show() { d.open = true }
