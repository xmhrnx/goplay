package app

import (
	"fmt"
	"math"

	"github.com/xmhrnx/giode/components/menu"

	"github.com/mehran9675/goplay/internal/player"
	"github.com/mehran9675/goplay/internal/settings"
)

// videoMenu changes how the picture is shown.
func (u *ui) videoMenu() *menu.Menu {
	a := u.a
	st := &a.player.State
	m := newMenu()
	if len(st.TracksOf(player.Video)) > 1 {
		tm := newMenu()
		u.trackMenu(tm, player.Video)
		m.SubMenu(tr("Video Track"), tm)
	}
	fit := newMenu()
	for _, o := range settings.Lookup("fit").Options {
		v := o.Value
		radio(fit, tr(o.Label), func() bool { return a.settings.Get("fit") == v }, func() { a.setSetting("fit", v) })
	}
	m.SubMenu(tr("Picture Size"), fit)
	aspect := newMenu()
	aspectOn := func(v string) func() bool {
		return func() bool {
			cur, _ := a.player.Get("video-aspect-override")
			return (v == "" && (cur == "no" || cur == "-1" || cur == "")) || aspectMatches(cur, v)
		}
	}
	for _, r := range []struct{ label, value string }{
		{"Original", ""}, {"16:9", "16:9"}, {"4:3", "4:3"}, {"21:9", "2.35:1"}, {"1.85:1", "1.85:1"}, {"1:1", "1:1"},
	} {
		v := r.value
		radio(aspect, tr(r.label), aspectOn(v), func() { a.player.SetAspect(v) })
	}
	m.SubMenu(tr("Aspect Ratio"), aspect)
	u.items(m, "-", "zoom_in", "zoom_out", "zoom_reset", "rotate")
	check(m, tr("Mirror"), func() bool { return u.flipH }, func() { u.flipH = !u.flipH; a.player.SetFlip(u.flipH, u.flipV) })
	check(m, tr("Flip Upside Down"), func() bool { return u.flipV }, func() { u.flipV = !u.flipV; a.player.SetFlip(u.flipH, u.flipV) })
	u.items(m, "-", "deinterlace", "-", "screenshot", "screenshot_clean", "-", "info")
	m.Separator()
	m.Item(tr("Video Settings…"), func() { u.dialogs.settings.show("Video") })
	return m
}

// aspectMatches reports whether mpv's aspect override cur is the ratio
// want (as "w:h").
func aspectMatches(cur, want string) bool {
	if want == "" {
		return false
	}
	var cw, ch, ww, wh float64
	if _, err := fmt.Sscanf(want, "%g:%g", &ww, &wh); err != nil || wh == 0 {
		return false
	}
	if _, err := fmt.Sscanf(cur, "%g:%g", &cw, &ch); err == nil && ch != 0 {
		return math.Abs(cw/ch-ww/wh) < 0.01
	}
	var f float64
	if _, err := fmt.Sscanf(cur, "%g", &f); err == nil {
		return math.Abs(f-ww/wh) < 0.01
	}
	return false
}
