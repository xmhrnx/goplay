package app

import (
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"

	"github.com/mehran9675/goplay/internal/settings"
)

func (v *videoArea) handleWheel(gtx layout.Context) {
	a := v.u.a
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &v.wheelTag, Kinds: pointer.Scroll, ScrollY: pointer.ScrollRange{Min: -1e6, Max: 1e6}})
		if !ok {
			return
		}
		e, ok := ev.(pointer.Event)
		if !ok || e.Scroll.Y == 0 || a.player.State.Idle {
			continue
		}
		steps := e.Scroll.Y / float32(gtx.Dp(48))
		switch {
		case a.vrActive():
			a.vrZoom(steps)
		case a.settings.Get(settings.KeyWheel) == "volume":
			a.addVolume(-float64(steps) * a.settings.Float(settings.KeyVolumeStep))
		case a.settings.Get(settings.KeyWheel) == "seek":
			a.seekBy(-float64(steps) * a.seekStep(false))
		}
	}
}

func (v *videoArea) handleVRDrag(gtx layout.Context) {
	a := v.u.a
	for {
		e, ok := v.vrDrag.Update(gtx.Metric, gtx.Source, gesture.Both)
		if !ok {
			return
		}
		switch e.Kind {
		case pointer.Press:
			v.vrLast = e.Position
			a.stereo.dragging = true
		case pointer.Drag:
			d := e.Position.Sub(v.vrLast)
			v.vrLast = e.Position
			a.vrDrag(d.X, d.Y)
		case pointer.Release, pointer.Cancel:
			a.vrDragEnd()
		}
	}
}
