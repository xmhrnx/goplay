package app

import (
	"image"
	"image/color"

	"gioui.org/layout"

	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/player"
)

// preview shows the time (and chapter) under the pointer above the bar.
func (s *seekBar) preview(gtx layout.Context, t float64, x float32) {
	label := player.FormatTime(t)
	if !s.u.a.casting() {
		st := &s.u.a.player.State
		if i := st.ChapterAt(t); i >= 0 && st.Chapters[i].Title != "" {
			label = st.Chapters[i].Title + " · " + label
		}
	}
	origin := elements.AbsolutePosition()
	anchor := image.Pt(origin.X+int(x)-gtx.Dp(40), origin.Y-gtx.Dp(38))
	box := elements.Box(styles.Styles{
		Background:   colSurface2,
		BorderRadius: 6,
		Padding:      properties.SymmetricInset(10, 5),
		BoxShadow:    properties.BoxShadow{Y: 2, Blur: 8, Color: "#00000099"},
	}, elements.Text(label, styles.Styles{Color: colText, FontSize: 12}))
	elements.RegisterPortal(gtx, anchor, func() elements.Element {
		// Portals get the whole window as their minimum; hug the label.
		return elements.Raw(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			return box.Layout(gtx)
		})
	})
}

// timeline is the position and duration to show: of the cast device
// while casting, of local playback otherwise.
func (a *App) timeline() (pos, dur float64) {
	if a.casting() {
		return a.cast.status.Position, a.cast.status.Duration
	}
	return a.player.State.Position, a.player.State.Duration
}

// seekFraction seeks to a fraction of the duration; exact lands on the
// exact frame.
func (a *App) seekFraction(f float64, exact bool) {
	if a.casting() {
		if exact {
			a.castSeek(f * a.cast.status.Duration)
		}
		return
	}
	a.player.SeekFraction(f, exact)
}

func rgba(rgb uint32, alpha uint8) color.NRGBA {
	return color.NRGBA{R: uint8(rgb >> 16), G: uint8(rgb >> 8), B: uint8(rgb), A: alpha}
}
