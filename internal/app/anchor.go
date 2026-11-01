package app

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"

	"github.com/mehran9675/giode/components/menu"
	"github.com/mehran9675/giode/elements"
)

// anchor records where an element was laid out, to open popovers from
// it.
type anchor struct {
	el elements.Element
	// pos is where it was in the last frame, exact once that frame was
	// laid out; tracked is the probe TrackPosition keeps moving.
	pos, tracked image.Point
	size         image.Point
}

func (an *anchor) Layout(gtx layout.Context) layout.Dimensions {
	// Rows place their children after laying them out; TrackPosition
	// follows them there. Clicks handled while el lays out read the
	// last frame's position, which is final.
	an.pos = an.tracked
	elements.TrackPosition(&an.tracked)
	d := an.el.Layout(gtx)
	an.size = d.Size
	return d
}

// openMenuAt shows m with its top-left at p (pixels); it flips above
// when there's no room below.
func (u *ui) openMenuAt(m *menu.Menu, p image.Point) {
	if u.menu != nil {
		u.menu.Close()
	}
	u.menu = m
	m.OpenAt(f32.Pt(float32(p.X), float32(p.Y)))
	u.a.dirty = true
}

// openMenuOver opens m from a control: above it when the control is in
// the bottom half of the window, below otherwise.
func (u *ui) openMenuOver(m *menu.Menu, an *anchor) {
	p := an.pos
	if p.Y > u.fb.Y/2 {
		u.openMenuAt(m, image.Pt(p.X, p.Y-u.dp(6)))
		return
	}
	u.openMenuAt(m, image.Pt(p.X, p.Y+an.size.Y+u.dp(6)))
}
