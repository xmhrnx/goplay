package app

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op"

	"github.com/xmhrnx/giode"
)

// drawVideo is the window's underlay: mpv draws the current video frame
// into the window's framebuffer, before the UI is drawn over it. It runs
// on the render thread with the window's OpenGL context current, where
// the render context is made and used.
func (a *App) drawVideo(t giode.GPUTarget) {
	if a.rend == nil {
		if a.rendFailed {
			return
		}
		rend, err := a.player.Handle().NewRenderContext(t.GetProcAddress)
		if err != nil {
			a.rendFailed = true
			a.appendLog("error", "render", err.Error())
			a.Post(func() {
				a.toastError(tr("Couldn't start video output: {error}", "error", err))
				a.runStarting()
			})
			return
		}
		rend.SetUpdateCallback(a.g.Invalidate)
		a.rend = rend
		a.Post(a.runStarting)
	}
	a.rend.Update()
	if !a.showVideo.Load() || t.Size.X <= 0 || t.Size.Y <= 0 {
		return
	}
	// GL's origin is the bottom left, the window's the top left.
	if err := a.rend.Render(int(t.Framebuffer), t.Size.X, t.Size.Y, true); err != nil {
		a.appendLog("error", "render", err.Error())
	}
}

// releaseVideo frees mpv's render context on the render thread, before
// the window's GL context goes away. If the context comes back,
// drawVideo makes a new one.
func (a *App) releaseVideo() {
	if a.rend != nil {
		a.rend.Free()
		a.rend = nil
	}
}

// updateVideoMargins keeps mpv's picture out of the title bar's and
// control bar's strips when they're beside the video (a normal window);
// maximized and fullscreen, the bars float over it.
func (a *App) updateVideoMargins(fb image.Point) {
	top, bottom := 0.0, 0.0
	if !a.ui.overlay() {
		top = float64(a.ui.titleHeightPx()) / float64(fb.Y)
		bottom = float64(a.ui.barHeightPx()) / float64(fb.Y)
		if top+bottom > 0.8 {
			top, bottom = 0, 0
		}
	}
	if top != a.ui.marginTop {
		a.player.Handle().SetDouble("video-margin-ratio-top", top)
		a.ui.marginTop = top
	}
	if bottom != a.ui.marginBottom {
		a.player.Handle().SetDouble("video-margin-ratio-bottom", bottom)
		a.ui.marginBottom = bottom
		// Subtitles go in the picture's black bars when there are any,
		// but the margins goplay adds are covered by its own bars.
		a.player.Handle().SetFlag("sub-use-margins", bottom == 0)
	}
}

// frame lays out one frame of the UI.
func (a *App) frame(gtx layout.Context) layout.Dimensions {
	fb := gtx.Constraints.Max
	a.ui.beginFrame(fb, gtx.Metric.PxPerDp)
	a.updateVideoMargins(fb)
	st := &a.player.State
	a.showVideo.Store(!st.Idle && !a.casting())
	a.ui.layout(gtx)

	if a.dirty {
		a.dirty = false
		gtx.Execute(op.InvalidateCmd{})
	} else if t := a.ui.nextTimer(); !t.IsZero() {
		gtx.Execute(op.InvalidateCmd{At: t})
	}
	return layout.Dimensions{Size: fb}
}
