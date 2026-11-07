package app

import (
	"net/url"
	"strconv"

	gioapp "gioui.org/app"
	"gioui.org/io/event"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/goplay/internal/settings"
)

// Post runs fn on the UI goroutine before the next frame. It is safe
// from any goroutine and doesn't wait.
func (a *App) Post(fn func()) { a.g.Post(fn) }

// processPlayer handles mpv's queued events.
func (a *App) processPlayer() {
	a.mpvPending.Store(false)
	if a.player.Process() {
		a.afterPlayerChange()
		a.dirty = true
	}
}

// windowEvent follows the window's size, for picture in picture and
// fitting the window to a video.
func (a *App) windowEvent(e event.Event) {
	if c, ok := e.(gioapp.ConfigEvent); ok {
		a.winSize = c.Config.Size
	}
}

// windowStateChanged pauses playback while the window is minimized, if
// the user asked for that, and resumes it after.
func (a *App) windowStateChanged(s giode.WindowState) {
	st := &a.player.State
	switch {
	case s == giode.WindowMinimized:
		if a.settings.Bool(settings.KeyPauseMinimized) && !st.Idle && !st.Paused && !a.casting() {
			a.player.SetPause(true)
			a.pausedHidden = true
		}
	case a.pausedHidden:
		a.pausedHidden = false
		a.player.SetPause(false)
	}
	a.dirty = true
}

// openedURL opens what macOS hands the app: files opened with it from
// Finder arrive as file:// URLs.
func (a *App) openedURL(u *url.URL) {
	item := u.String()
	if u.Scheme == "file" {
		item = u.Path
	}
	a.whenReady(func() { a.openItems([]string{item}, false) })
}

// whenReady runs fn now if the video output is ready, or once it is.
func (a *App) whenReady(fn func()) {
	if a.ready {
		fn()
		return
	}
	a.starting = append(a.starting, fn)
}

// runStarting runs what waited for the video output.
func (a *App) runStarting() {
	a.ready = true
	fns := a.starting
	a.starting = nil
	for _, fn := range fns {
		fn()
	}
	a.dirty = true
}

// shutdown saves state once the window has closed. mpv is left to the
// process's exit: its render context belonged to the window's GL
// context, which is gone, and the core mustn't outlive it.
func (a *App) shutdown() {
	st := &a.player.State
	if a.lastPath != "" && st.Duration > 0 {
		a.history.Update(a.lastPath, a.lastTitle, st.Position, st.Duration)
	}
	a.history.Save()
	if a.settings.Bool(settings.KeyRememberVolume) && st.Volume > 0 {
		a.settings.Set(settings.KeyVolume, strconv.FormatFloat(st.Volume, 'f', 0, 64))
	}
	a.settings.Flush()
	a.stopCasting()
	if !a.resumeSaved {
		a.player.SaveResumePosition()
	}
}

// runCommands runs commands by ID (from --run).
func (a *App) runCommands(ids []string) {
	for _, id := range ids {
		if a.actions.Get(id) == nil {
			a.toastError(tr("Unknown command: {command}", "command", id))
			continue
		}
		a.actions.Run(id)
	}
	a.dirty = true
}

// Quit closes the window, which ends the app. Playback stops first, so
// nothing is heard while the window and the player shut down.
func (a *App) Quit() {
	a.player.SaveResumePosition()
	a.resumeSaved = true
	a.player.SetPause(true)
	a.g.Close()
}
