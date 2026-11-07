package app

import (
	"path/filepath"
	"time"

	"github.com/mehran9675/goplay/internal/player"
	"github.com/mehran9675/goplay/internal/settings"
)

// stereoState is how the current file's 3D or 360° picture is shown.
type stereoState struct {
	layout     player.StereoLayout
	view       player.StereoView
	projection player.Projection
	vrStereo   player.VRStereo
	vr         player.VRView
	// outW and outH are the 360° view's size in pixels.
	outW, outH int
	// pending marks a view change not yet sent to mpv (drags are sent at
	// most every vrInterval).
	pending  bool
	lastSent time.Time
	// live is whether the 360° filter takes angle changes as commands;
	// otherwise it's rebuilt for each change.
	live bool
	// dragging renders the 360° view at reduced size while the user
	// drags, so rebuilding it keeps up; full size returns on release.
	dragging bool
}

// vrInterval paces 360° view updates while dragging.
const vrInterval = 40 * time.Millisecond

func (a *App) vrActive() bool { return a.stereo.projection != player.ProjectionFlat }

func (a *App) stereoActive() bool { return a.stereo.layout != player.Layout2D }

// stereoFileLoaded resets the picture for a new file and, if enabled,
// recognizes 3D and 360° files by name.
func (a *App) stereoFileLoaded() {
	s := &a.stereo
	*s = stereoState{vr: player.DefaultVRView}
	s.vr.FOV = a.settings.Float(settings.KeyVRFOV)
	if a.settings.Bool(settings.KeyAuto3D) {
		st := &a.player.State
		d := player.DetectStereo(filepath.Base(st.Path)+" "+st.Title, st.VideoW, st.VideoH)
		switch {
		case d.Projection != player.ProjectionFlat:
			s.projection, s.vrStereo = d.Projection, d.VRStereo
			a.toast(tr("360° video — drag to look around, scroll to zoom"))
		case d.Layout != player.Layout2D:
			s.layout = d.Layout
			s.view = player.StereoView(a.settings.Get(settings.KeyStereoView))
			a.toast(tr("3D video — change how it's shown in the 3D menu"))
		}
	}
	a.player.SetFilter("goplay-3d", "")
	a.player.SetFilter("goplay-360", "")
	a.applyStereo()
}

// setStereoLayout says how the file packs its two views; Layout2D turns
// 3D handling off.
func (a *App) setStereoLayout(l player.StereoLayout) {
	a.stereo.layout = l
	a.stereo.projection = player.ProjectionFlat
	if a.stereo.view == "" {
		a.stereo.view = player.StereoView(a.settings.Get(settings.KeyStereoView))
	}
	a.applyStereo()
	if l == player.Layout2D {
		a.osd(tr("3D off"))
	} else {
		a.osd(tr("3D on"))
	}
}

// setStereoView chooses how 3D is shown (anaglyph, one eye, ...).
func (a *App) setStereoView(v player.StereoView) {
	a.stereo.view = v
	if a.stereo.layout == player.Layout2D {
		a.stereo.layout = player.LayoutSBSHalf
	}
	a.applyStereo()
}

// setProjection treats the file as 360° or 180° footage (or flat).
func (a *App) setProjection(p player.Projection, s player.VRStereo) {
	a.stereo.projection, a.stereo.vrStereo = p, s
	a.stereo.layout = player.Layout2D
	if a.stereo.vr.FOV == 0 {
		a.stereo.vr = player.DefaultVRView
		a.stereo.vr.FOV = a.settings.Float(settings.KeyVRFOV)
	}
	a.applyStereo()
	if p == player.ProjectionFlat {
		a.osd(tr("Flat video"))
	} else {
		a.osd(tr("Drag to look around · scroll to zoom"))
	}
}

func (a *App) resetVRView() {
	a.stereo.vr = player.DefaultVRView
	a.stereo.vr.FOV = a.settings.Float(settings.KeyVRFOV)
	a.applyStereo()
}

// applyStereo installs the 3D or 360° filter for the current state.
func (a *App) applyStereo() {
	s := &a.stereo
	if err := a.player.SetFilter("goplay-3d", player.StereoFilter(s.layout, s.view)); err != nil {
		a.toastError(tr("3D: {error}", "error", err))
	}
	if s.projection == player.ProjectionFlat {
		a.player.SetFilter("goplay-360", "")
		return
	}
	w, h := a.vrOutputSize()
	s.outW, s.outH = w, h
	if err := a.player.SetFilter("goplay-360", player.VRFilter(s.projection, s.vrStereo, s.vr, w, h)); err != nil {
		a.toastError(tr("360°: {error}", "error", err))
		return
	}
	s.live = true
	s.pending = false
	s.lastSent = time.Now()
}
