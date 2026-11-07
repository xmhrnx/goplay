package app

import (
	"image"

	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"

	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles/properties"
)

// layout draws the whole UI.
func (u *ui) layout(gtx layout.Context) {
	fb := gtx.Constraints.Max
	overlay := u.overlay()
	titleH, barH := u.titleHeightPx(), u.barHeightPx()
	st := &u.a.player.State

	// Overlay mode: hovering a strip shows its bar. Menus and drags
	// keep the bottom bar up; with nothing loaded both bars stay.
	topZone := image.Rect(0, 0, fb.X, titleH+u.dp(16))
	bottomZone := image.Rect(0, fb.Y-barH-u.dp(8), fb.X, fb.Y)
	topHovered := u.topHover.Update(gtx.Source)
	bottomHovered := u.bottomHover.Update(gtx.Source)
	u.topShown = !overlay || st.Idle || topHovered
	u.bottomShown = !overlay || st.Idle || bottomHovered || u.bar.busy() || (u.menu != nil && u.menu.Opened())
	if u.a.g.Fullscreen() && !st.Idle {
		u.topShown = u.topShown && topHovered
	}
	if st.Idle || u.a.casting() {
		// mpv draws nothing then; the window shows through.
		paint.Fill(gtx.Ops, properties.CalcColor(colBackground))
	}

	// The video area, with everything drawn over the picture.
	place(gtx, u.videoRect, u.video)

	// The bars.
	if u.topShown {
		u.title.Title(u.windowTitle())
		place(gtx, image.Rect(0, 0, fb.X, titleH), u.titleBar(overlay))
	}
	if u.bottomShown {
		place(gtx, image.Rect(0, fb.Y-barH, fb.X, fb.Y), u.bar.element(overlay, fb.X))
	}
	if overlay {
		// Hover zones go on top and pass events through, so moving onto
		// a bar's buttons still counts as hovering its strip.
		zone(gtx, topZone, &u.topHover)
		zone(gtx, bottomZone, &u.bottomHover)
	}

	u.drawer.Layout(gtx)
	u.dialogs.layout(gtx)
	if u.menu != nil {
		u.menu.Layout(gtx)
	}
	u.toasts.Layout(gtx)
	u.trackPointer(gtx)
}

// place lays el out exactly filling r, with AbsolutePosition kept in
// step so popups inside it open in the right spot.
func place(gtx layout.Context, r image.Rectangle, el elements.Element) layout.Dimensions {
	if r.Empty() {
		return layout.Dimensions{}
	}
	t := op.Offset(r.Min).Push(gtx.Ops)
	pop := elements.PushOffset(r.Min)
	cgtx := gtx
	cgtx.Constraints = layout.Exact(r.Size())
	dims := el.Layout(cgtx)
	pop()
	t.Pop()
	return dims
}

// zone registers a pass-through hover area.
func zone(gtx layout.Context, r image.Rectangle, h *gesture.Hover) {
	area := clip.Rect(r).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	h.Add(gtx.Ops)
	pass.Pop()
	area.Pop()
}
