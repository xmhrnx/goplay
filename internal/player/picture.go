package player

import (
	"strconv"
	"strings"
)

// Picture.

// Screenshot saves a screenshot to the screenshot folder: "subtitles"
// includes subtitles as shown, "video" the bare frame, "window" the
// scaled picture.
func (p *Player) Screenshot(mode string) error { return p.h.Command("screenshot", mode) }

// SetAspect forces a display aspect ratio ("16:9"); "" uses the file's.
func (p *Player) SetAspect(ratio string) {
	if ratio == "" {
		ratio = "no"
	}
	p.h.SetString("video-aspect-override", ratio)
}

// Fit modes for SetFit.
const (
	FitContain = "contain" // whole picture, letterboxed
	FitCover   = "cover"   // fill the window, cropping
	FitStretch = "stretch" // fill the window, distorting
)

// SetFit chooses how the picture fills the window.
func (p *Player) SetFit(mode string) {
	switch mode {
	case FitCover:
		p.h.SetString("keepaspect", "yes")
		p.h.SetDouble("panscan", 1)
	case FitStretch:
		p.h.SetString("keepaspect", "no")
		p.h.SetDouble("panscan", 0)
	default:
		p.h.SetString("keepaspect", "yes")
		p.h.SetDouble("panscan", 0)
	}
}

// SetZoom zooms the picture: 0 is none, 1 doubles it (log2 scale).
func (p *Player) SetZoom(z float64) { p.h.SetDouble("video-zoom", z) }

// SetPan moves the zoomed picture, in fractions of its size.
func (p *Player) SetPan(x, y float64) {
	p.h.SetDouble("video-pan-x", x)
	p.h.SetDouble("video-pan-y", y)
}

// SetRotation rotates the picture by a multiple of 90 degrees.
func (p *Player) SetRotation(deg int) {
	p.h.SetString("video-rotate", strconv.Itoa(((deg%360)+360)%360))
}

// SetFlip mirrors the picture horizontally and/or vertically.
func (p *Player) SetFlip(h, v bool) {
	var f []string
	if h {
		f = append(f, "hflip")
	}
	if v {
		f = append(f, "vflip")
	}
	filter := ""
	if len(f) > 0 {
		filter = "lavfi=[" + strings.Join(f, ",") + "]"
	}
	p.SetFilter("goplay-flip", filter)
}
