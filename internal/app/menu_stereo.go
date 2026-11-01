package app

import (
	"github.com/mehran9675/giode/components/menu"

	"github.com/mehran9675/goplay/internal/player"
	"github.com/mehran9675/goplay/internal/settings"
)

// stereoMenu chooses how 3D and 360° video is treated.
func (u *ui) stereoMenu() *menu.Menu {
	a := u.a
	s := &a.stereo
	m := newMenu()
	radio(m, tr("Normal (2D)"), func() bool { return s.layout == player.Layout2D && s.projection == player.ProjectionFlat }, func() {
		a.setProjection(player.ProjectionFlat, "")
		a.setStereoLayout(player.Layout2D)
	})

	in := newMenu()
	for _, l := range []struct {
		label  string
		layout player.StereoLayout
	}{
		{"Side by side (half width)", player.LayoutSBSHalf},
		{"Side by side (full width)", player.LayoutSBSFull},
		{"Side by side, right eye first", player.LayoutSBSHalfRight},
		{"Top and bottom (half height)", player.LayoutTBHalf},
		{"Top and bottom (full height)", player.LayoutTBFull},
		{"Top and bottom, right eye first", player.LayoutTBHalfRight},
	} {
		l := l
		radio(in, tr(l.label), func() bool { return s.layout == l.layout }, func() { a.setStereoLayout(l.layout) })
	}
	m.SubMenu(tr("3D Video Format"), in)

	out := newMenu()
	for _, o := range settings.StereoViews {
		v := player.StereoView(o.Value)
		radio(out, tr(o.Label), func() bool { return s.layout != player.Layout2D && s.view == v }, func() { a.setStereoView(v) })
	}
	m.SubMenu(tr("Show 3D As"), out)
	m.Separator()

	vr := newMenu()
	for _, p := range []struct {
		label  string
		proj   player.Projection
		stereo player.VRStereo
	}{
		{"360° video", player.ProjectionEquirect, player.VRMono},
		{"360° 3D, top and bottom", player.ProjectionEquirect, player.VRTopBottom},
		{"360° 3D, side by side", player.ProjectionEquirect, player.VRSideBySide},
		{"180° 3D (VR180)", player.ProjectionHalf, player.VRSideBySide},
		{"180° video", player.ProjectionHalf, player.VRMono},
		{"Fisheye 180° 3D", player.ProjectionFisheye, player.VRSideBySide},
		{"YouTube cubemap (EAC)", player.ProjectionEAC, player.VRMono},
	} {
		p := p
		radio(vr, tr(p.label), func() bool { return s.projection == p.proj && s.vrStereo == p.stereo }, func() { a.setProjection(p.proj, p.stereo) })
	}
	m.SubMenu(tr("360° and VR"), vr)
	u.items(m, "vr_reset")
	m.Separator()
	m.Item(tr("3D Settings…"), func() { u.dialogs.settings.show("3D & VR") })
	return m
}
