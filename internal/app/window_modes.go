package app

import (
	"image"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/goplay/internal/settings"
)

// Window modes.

// pipWidth is the picture-in-picture window's width in dp, and pipMargin
// its distance from the screen's edges.
const pipWidth, pipMargin = 420, 16

// maxFit is the share of the screen's work area sizeToVideo fills at
// most, and maxFitW, maxFitH the limit in dp where the screen's size
// isn't known (Wayland).
const (
	maxFit           = 0.9
	maxFitW, maxFitH = 1280, 800
)

func (a *App) setFullscreen(on bool) {
	if a.ui.pip && on {
		a.togglePiP()
	}
	a.g.SetFullscreen(on)
	a.dirty = true
}

// setOnTop keeps the window above others, or not.
func (a *App) setOnTop(on bool) {
	a.g.SetAlwaysOnTop(on)
}

// togglePiP shrinks the window to a small always-on-top player in the
// screen's bottom right corner, or brings it back.
func (a *App) togglePiP() {
	u := a.ui
	if u.pip {
		u.pip = false
		a.setOnTop(u.pipWasOnTop)
		a.g.SetSize(int(u.pipRestore.Width+0.5), int(u.pipRestore.Height+0.5))
		if u.pipHadFrame {
			a.g.SetPosition(u.pipRestore.X, u.pipRestore.Y)
		}
		return
	}
	if a.g.Fullscreen() {
		a.g.SetFullscreen(false)
	}
	if a.g.Maximized() {
		a.g.Unmaximize()
	}
	u.pip = true
	sz := a.windowSizeDp()
	u.pipRestore = giode.Rect{Width: float32(sz.X), Height: float32(sz.Y)}
	if f, err := a.g.Frame(); err == nil {
		u.pipRestore.X, u.pipRestore.Y = f.X, f.Y
		u.pipHadFrame = true
	} else {
		u.pipHadFrame = false
	}
	u.pipWasOnTop = a.g.AlwaysOnTop()
	w, h := pipWidth, pipWidth*9/16
	if st := &a.player.State; st.VideoW > 0 && st.VideoH > 0 {
		h = w * st.VideoH / st.VideoW
	}
	a.setOnTop(true)
	a.g.SetSize(w, h)
	// Where windows can't be placed (Wayland), it stays where it is.
	a.g.MoveTo(giode.CornerBottomRight, pipMargin)
}

// windowSizeDp is the window's size in dp.
func (a *App) windowSizeDp() image.Point {
	s := a.ui.scale
	if s <= 0 {
		s = 1
	}
	return image.Pt(int(float32(a.winSize.X)/s+0.5), int(float32(a.winSize.Y)/s+0.5))
}

// sizeToVideo fits the window to the video when a new one starts,
// leaving room for the bars around it.
func (a *App) sizeToVideo() {
	st := &a.player.State
	if !a.settings.Bool(settings.KeyResizeToVideo) || st.VideoW <= 0 || st.VideoH <= 0 ||
		a.g.Maximized() || a.g.Fullscreen() || a.ui.pip {
		return
	}
	// One video pixel per screen pixel, in dp.
	s := a.ui.scale
	if s <= 0 {
		s = 1
	}
	chrome := a.ui.titleHeightDp() + controlBarH
	maxW, maxH := float32(maxFitW), float32(maxFitH)
	if sc, err := a.g.Screen(); err == nil && sc.WorkArea.Width > 0 {
		maxW, maxH = sc.WorkArea.Width*maxFit, sc.WorkArea.Height*maxFit
	}
	w, h := float32(st.VideoW)/s, float32(st.VideoH)/s
	if f := min(maxW/w, (maxH-float32(chrome))/h, 1); f < 1 {
		w, h = w*f, h*f
	}
	a.g.SetSize(max(int(w), 480), max(int(h), 270)+chrome)
}
