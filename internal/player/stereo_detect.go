package player

import "strings"

// Detected is what a file's name (and picture shape) suggests about it.
type Detected struct {
	Layout     StereoLayout
	Projection Projection
	VRStereo   VRStereo
}

// DetectStereo guesses a file's 3D layout or 360° projection from the
// usual naming conventions ("Movie.3D.HSBS.mkv", "clip_360_TB.mp4",
// "VR180") and its size: a side-by-side picture twice as wide as usual
// is full width, otherwise half.
func DetectStereo(name string, w, h int) Detected {
	tokens := map[string]bool{}
	for _, t := range strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	}) {
		tokens[t] = true
	}
	has := func(ts ...string) bool {
		for _, t := range ts {
			if tokens[t] {
				return true
			}
		}
		return false
	}
	aspect := 0.0
	if h > 0 {
		aspect = float64(w) / float64(h)
	}

	sbs := has("sbs", "hsbs", "halfsbs", "fsbs", "fullsbs", "lr", "3dsbs")
	tb := has("tb", "tab", "htab", "ou", "hou", "overunder", "topbottom", "ftab", "fou", "3dou", "3dtab")

	switch {
	case has("180", "vr180", "180vr", "180x180"):
		d := Detected{Projection: ProjectionHalf, VRStereo: VRSideBySide}
		if tb {
			d.VRStereo = VRTopBottom
		}
		if has("fisheye") {
			d.Projection = ProjectionFisheye
		}
		return d
	case has("360", "vr360", "360vr", "equirect", "equirectangular", "eac"):
		d := Detected{Projection: ProjectionEquirect, VRStereo: VRMono}
		if has("eac") {
			d.Projection = ProjectionEAC
		}
		switch {
		case tb || (aspect > 0 && aspect < 1.5 && has("3d", "stereo")):
			d.VRStereo = VRTopBottom
		case sbs:
			d.VRStereo = VRSideBySide
		}
		return d
	case sbs:
		if has("fsbs", "fullsbs") || aspect >= 3 {
			return Detected{Layout: LayoutSBSFull}
		}
		return Detected{Layout: LayoutSBSHalf}
	case tb:
		if has("ftab", "fou") || (aspect > 0 && aspect <= 1) {
			return Detected{Layout: LayoutTBFull}
		}
		return Detected{Layout: LayoutTBHalf}
	}
	return Detected{}
}
