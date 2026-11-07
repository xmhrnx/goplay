package app

import (
	"time"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
)

// trackPointer follows the pointer over the whole window, without
// taking its events from anything: where it is (for the context menu)
// and when it last moved. Over playing video in overlay mode it hides
// once it rests for a moment, as in every player.
func (u *ui) trackPointer(gtx layout.Context) {
	area := clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, &u.pointerTag)
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &u.pointerTag, Kinds: pointer.Move | pointer.Drag | pointer.Press | pointer.Scroll})
		if !ok {
			break
		}
		if e, ok := ev.(pointer.Event); ok {
			u.lastPointer = e.Position
			u.lastMove = time.Now()
		}
	}
	st := &u.a.player.State
	u.cursorHidden = u.overlay() && !st.Idle && !st.Paused && !u.topShown && !u.bottomShown && !u.modalOpen() &&
		time.Since(u.lastMove) >= u.hideDelay()
	if u.cursorHidden {
		pointer.CursorNone.Add(gtx.Ops)
	}
	pass.Pop()
	area.Pop()
}
