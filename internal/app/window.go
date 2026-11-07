package app

// windowCtl lets the title bar drive the window.
type windowCtl struct{ a *App }

func (w windowCtl) Minimize() { w.a.g.Minimize() }
func (w windowCtl) ToggleMaximize() {
	if w.a.ui.pip {
		w.a.togglePiP()
	}
	if w.a.g.Fullscreen() {
		w.a.setFullscreen(false)
		return
	}
	w.a.g.ToggleMaximize()
}
func (w windowCtl) Close()          { w.a.Quit() }
func (w windowCtl) Maximized() bool { return w.a.g.Maximized() || w.a.g.Fullscreen() }
func (w windowCtl) Focused() bool   { return w.a.g.Focused() }
