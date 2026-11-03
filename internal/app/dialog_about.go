package app

import (
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"
)

// aboutDialog shows GoPlay's version and what it's built with.
type aboutDialog struct {
	modal
}

func newAboutDialog(u *ui) *aboutDialog {
	a := u.a
	d := &aboutDialog{}
	site := secondaryButton(tr("Website"), func() { openWebPage("https://github.com/Mehran9675/goplay") })
	closeBtn := primaryButton(tr("Close"), func() { d.open = false })
	d.dlg = dialog.New(&d.open, func() elements.Element {
		mpvVersion, _ := a.player.Get("mpv-version")
		ffVersion, _ := a.player.Get("ffmpeg-version")
		return column(10,
			elements.Row(styles.Styles{Gap: 12, Align: properties.AlignCenter}, ic("logo", 40, colHighlight), heading("GoPlay")),
			note(tr("Version {version}", "version", Version)),
			note(tr("Plays nearly anything with {mpv} and FFmpeg {ffmpeg}.", "mpv", mpvVersion, "ffmpeg", ffVersion)),
			note(tr("Interface built with Giode and Gio. Icons from Material Design.")),
			buttonsRow(site, closeBtn))
	}, dialogStyles(440))
	return d
}

func (d *aboutDialog) show() { d.open = true }
