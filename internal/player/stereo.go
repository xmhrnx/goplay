package player

import (
	"fmt"
	"math"
)

// 3D and VR video. Both are done with FFmpeg filters in mpv's filter
// chain: stereo3d converts between stereoscopic layouts and outputs
// (anaglyph, one eye, interleaved, ...), v360 projects 360° and 180°
// footage onto a flat view that the user steers.

// StereoLayout is how a stereoscopic file packs its two views (stereo3d
// input names).
type StereoLayout string

// StereoView is how to show stereoscopic video (stereo3d output names).
type StereoView string

// Projection is how a 360° or 180° file maps the sphere (v360 input
// names).
type Projection string

// VRStereo is how a stereoscopic 360°/180° file packs its eyes (v360
// in_stereo names).
type VRStereo string

// VRView is where the viewer looks, in degrees: yaw turns right, pitch
// looks up, FOV is the horizontal field of view.
type VRView struct {
	Yaw, Pitch, Roll, FOV float64
}

const (
	Layout2D            StereoLayout = ""
	LayoutSBSHalf       StereoLayout = "sbs2l" // side by side, squeezed, left eye first
	LayoutSBSHalfRight  StereoLayout = "sbs2r" // right eye first (cross-eyed)
	LayoutSBSFull       StereoLayout = "sbsl"
	LayoutSBSFullRight  StereoLayout = "sbsr"
	LayoutTBHalf        StereoLayout = "ab2l" // top and bottom, squeezed, left eye on top
	LayoutTBHalfRight   StereoLayout = "ab2r"
	LayoutTBFull        StereoLayout = "abl"
	LayoutTBFullRight   StereoLayout = "abr"
	LayoutAlternateLeft StereoLayout = "al" // frame sequential
)

const (
	ViewLeft            StereoView = "ml" // 2D, left eye
	ViewRight           StereoView = "mr" // 2D, right eye
	ViewRedCyan         StereoView = "arcd"
	ViewRedCyanColor    StereoView = "arcc"
	ViewGreenMagenta    StereoView = "agmd"
	ViewYellowBlue      StereoView = "aybd"
	ViewSideBySide      StereoView = "sbs2l" // half side by side, for 3D TVs
	ViewTopBottom       StereoView = "ab2l"
	ViewInterleavedRows StereoView = "irl" // passive (polarized) 3D monitors
	ViewInterleavedCols StereoView = "icl"
	ViewCheckerboard    StereoView = "chl" // DLP 3D projectors
)

const (
	ProjectionFlat     Projection = ""
	ProjectionEquirect Projection = "e"         // 360° equirectangular
	ProjectionHalf     Projection = "hequirect" // 180° (VR180)
	ProjectionFisheye  Projection = "fisheye"
	ProjectionEAC      Projection = "eac" // equi-angular cubemap (YouTube)
	ProjectionCubemap  Projection = "c3x2"
)

const (
	VRMono       VRStereo = "2d"
	VRSideBySide VRStereo = "sbs"
	VRTopBottom  VRStereo = "tb"
)

// DefaultVRView looks straight ahead with a natural field of view.
var DefaultVRView = VRView{FOV: 100}

// StereoFilter returns the mpv video filter that shows a file packed as
// layout the way view asks, or "" when none is needed.
func StereoFilter(layout StereoLayout, view StereoView) string {
	if layout == Layout2D || view == "" || string(view) == string(layout) {
		return ""
	}
	return fmt.Sprintf("lavfi=[stereo3d=in=%s:out=%s]", layout, view)
}

// Clamp keeps the view in range: yaw wraps, pitch stops at the poles,
// FOV stays between a narrow zoom and a wide angle.
func (v VRView) Clamp() VRView {
	v.Yaw = math.Mod(v.Yaw+540, 360) - 180
	v.Pitch = math.Max(-90, math.Min(90, v.Pitch))
	v.Roll = math.Mod(v.Roll+540, 360) - 180
	v.FOV = math.Max(30, math.Min(150, v.FOV))
	return v
}

// VerticalFOV is the vertical field of view for an output of w×h pixels.
func (v VRView) VerticalFOV(w, h int) float64 {
	if w <= 0 || h <= 0 {
		return v.FOV
	}
	half := v.FOV * math.Pi / 360
	return 2 * math.Atan(math.Tan(half)*float64(h)/float64(w)) * 180 / math.Pi
}

// VRFilter returns the mpv video filter projecting proj footage onto a
// flat w×h view, or "" for flat video.
func VRFilter(proj Projection, stereo VRStereo, view VRView, w, h int) string {
	if proj == ProjectionFlat {
		return ""
	}
	if stereo == "" {
		stereo = VRMono
	}
	view = view.Clamp()
	return fmt.Sprintf("lavfi=[v360=input=%s:output=flat:in_stereo=%s:out_stereo=2d:"+
		"yaw=%.2f:pitch=%.2f:roll=%.2f:h_fov=%.2f:v_fov=%.2f:w=%d:h=%d:interp=line]",
		proj, stereo, view.Yaw, view.Pitch, view.Roll, view.FOV, view.VerticalFOV(w, h), even(w), even(h))
}

// even rounds down to an even size, which video filters need.
func even(n int) int {
	if n < 2 {
		return 2
	}
	return n &^ 1
}
