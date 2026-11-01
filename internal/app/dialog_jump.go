package app

import (
	"strconv"
	"strings"

	"github.com/mehran9675/giode/components/dialog"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/elements"

	"github.com/mehran9675/goplay/internal/player"
)

// jumpDialog asks for a time to go to.
type jumpDialog struct {
	modal
	u     *ui
	input *input.Input
}

func newJumpDialog(u *ui) *jumpDialog {
	d := &jumpDialog{u: u}
	d.input = input.New("1:23:45, 83, or 50%", fieldStyles()).
		Submit(func(s string) { d.jumpTo(s) })
	goBtn := primaryButton(tr("Go"), func() { d.jumpTo(d.input.Text()) })
	cancel := secondaryButton(tr("Cancel"), func() { d.open = false })
	d.dlg = dialog.New(&d.open, func() elements.Element {
		return column(12, heading(tr("Go to Time")), d.input, buttonsRow(cancel, goBtn))
	}, dialogStyles(360))
	return d
}

func (d *jumpDialog) show() {
	d.input.SetText(player.FormatTime(d.u.a.player.State.Position))
	d.open = true
	d.input.Focus()
	d.input.SelectAll()
}

// jumpTo seeks to a time typed as h:mm:ss, m:ss, seconds or a percentage.
func (d *jumpDialog) jumpTo(s string) {
	a := d.u.a
	t, ok := parseTimeInput(s, a.player.State.Duration)
	if !ok {
		a.toastError(tr("That isn't a time: try 1:23:45, 83 or 50%"))
		return
	}
	d.open = false
	a.seekTo(t)
}

func parseTimeInput(s string, duration float64) (float64, bool) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "%") {
		f, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		if err != nil || duration <= 0 {
			return 0, false
		}
		return duration * f / 100, true
	}
	var total float64
	for _, part := range strings.Split(s, ":") {
		f, err := strconv.ParseFloat(strings.TrimSpace(part), 64)
		if err != nil || f < 0 {
			return 0, false
		}
		total = total*60 + f
	}
	return total, s != ""
}
