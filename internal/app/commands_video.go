package app

import "math"

func (a *App) screenshot(mode string) {
	if err := a.player.Screenshot(mode); err != nil {
		a.toastError(tr("Couldn't take a screenshot: {error}", "error", err))
		return
	}
	dir, _ := a.player.Get("screenshot-dir")
	a.toast(tr("Screenshot saved to {folder}", "folder", dir))
}

// Picture transforms, remembered for this file only.

func (a *App) rotate() {
	cur, _ := a.player.Handle().Int("video-rotate")
	next := int(cur+90) % 360
	a.player.SetRotation(next)
	a.osd(tr("Rotation {degrees}°", "degrees", next))
}

func (a *App) zoomBy(d float64) {
	cur, _ := a.player.Handle().Double("video-zoom")
	v := math.Max(-2, math.Min(3, cur+d))
	a.player.SetZoom(v)
	a.osd(tr("Zoom {percent}%", "percent", int(math.Round(math.Pow(2, v)*100))))
}

func (a *App) resetZoom() {
	a.player.SetZoom(0)
	a.player.SetPan(0, 0)
	a.osd(tr("Zoom 100%"))
}

func (a *App) cycleFit() {
	order := []string{"contain", "cover", "stretch"}
	names := map[string]string{"contain": tr("Fit the window"), "cover": tr("Fill the window"), "stretch": tr("Stretch")}
	cur := a.settings.Get("fit")
	next := order[0]
	for i, f := range order {
		if f == cur {
			next = order[(i+1)%len(order)]
		}
	}
	a.setSetting("fit", next)
	a.osd(names[next])
}

// toggleSetting flips a Toggle setting and applies it.
func (a *App) toggleSetting(key, label string) {
	on := !a.settings.Bool(key)
	v := "no"
	if on {
		v = "yes"
	}
	a.setSetting(key, v)
	a.osd(onOff(tr(label), on))
}

func (a *App) toggleOnTop() {
	on := !a.g.AlwaysOnTop()
	a.setOnTop(on)
	a.osd(onOff(tr("Always on top"), on))
}
