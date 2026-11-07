package app

import (
	"math"
	"strconv"
	"time"

	"github.com/mehran9675/goplay/internal/player"
	"github.com/mehran9675/goplay/internal/settings"
)

// vrOutputSize is the 360° view's pixel size: the video area, capped so
// the projection stays cheap.
func (a *App) vrOutputSize() (int, int) {
	r := a.ui.videoRect
	w, h := r.Dx(), r.Dy()
	if w <= 0 || h <= 0 {
		w, h = 1920, 1080
	}
	if w > 1920 {
		h = h * 1920 / w
		w = 1920
	}
	if a.stereo.dragging {
		w, h = w/2, h/2
	}
	return w, h
}

// vrDrag turns the view by a pointer drag of dx, dy pixels.
func (a *App) vrDrag(dx, dy float32) {
	s := &a.stereo
	w := float32(max(a.ui.videoRect.Dx(), 1))
	// One video-width of drag turns by the field of view.
	speed := float64(s.vr.FOV) / float64(w) * a.settings.Float(settings.KeyVRSensitivity)
	sign := -1.0
	if a.settings.Bool(settings.KeyVRInvert) {
		sign = 1
	}
	s.vr.Yaw += sign * float64(dx) * speed
	s.vr.Pitch -= sign * float64(dy) * speed
	s.vr = s.vr.Clamp()
	a.sendVRView()
}

// vrZoom narrows (positive) or widens the field of view.
func (a *App) vrZoom(steps float32) {
	s := &a.stereo
	s.vr.FOV *= math.Pow(0.92, float64(steps))
	s.vr = s.vr.Clamp()
	a.sendVRView()
	a.osd(tr("Field of view {degrees}°", "degrees", int(math.Round(s.vr.FOV))))
}

// sendVRView updates the 360° filter's angles, no more often than
// vrInterval; the last change is sent when the timer fires.
func (a *App) sendVRView() {
	s := &a.stereo
	if time.Since(s.lastSent) < vrInterval {
		s.pending = true
		return
	}
	s.pending = false
	s.lastSent = time.Now()
	if s.live && !s.dragging {
		vfov := s.vr.VerticalFOV(s.outW, s.outH)
		ok := true
		for _, c := range [][2]string{
			{"yaw", ftoa(s.vr.Yaw)}, {"pitch", ftoa(s.vr.Pitch)},
			{"h_fov", ftoa(s.vr.FOV)}, {"v_fov", ftoa(vfov)},
		} {
			if a.player.FilterCommand("goplay-360", c[0], c[1]) != nil {
				ok = false
				break
			}
		}
		if ok {
			return
		}
		// This FFmpeg can't change the angles live: rebuild instead.
		s.live = false
	}
	w, h := a.vrOutputSize()
	a.player.SetFilter("goplay-360", player.VRFilter(s.projection, s.vrStereo, s.vr, w, h))
}

// vrDragEnd restores full resolution after a drag.
func (a *App) vrDragEnd() {
	if !a.stereo.dragging {
		return
	}
	a.stereo.dragging = false
	a.stereo.pending = false
	a.applyStereo()
}

// vrTimer is when a paced view change is due.
func (a *App) vrTimer() time.Time {
	if !a.stereo.pending {
		return time.Time{}
	}
	return a.stereo.lastSent.Add(vrInterval)
}

// vrTick sends a pending view change once due.
func (a *App) vrTick() {
	if a.stereo.pending && time.Since(a.stereo.lastSent) >= vrInterval {
		a.sendVRView()
	}
}

func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', 2, 64) }
