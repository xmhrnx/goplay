package app

import (
	"time"

	"github.com/xmhrnx/giode/components/toaster"

	"github.com/mehran9675/goplay/internal/settings"
)

// On-screen display.

const (
	osdDuration   = 1400 * time.Millisecond
	flashDuration = 500 * time.Millisecond
)

// osd briefly shows a message over the video.
func (a *App) osd(text string) { a.osdLevel(text, -1) }

// osdLevel shows a message with a level bar (volume), level from 0 to 1.
func (a *App) osdLevel(text string, level float64) {
	if !a.settings.Bool(settings.KeyOSD) {
		return
	}
	u := a.ui
	u.osdText, u.osdLevel, u.osdUntil = text, level, time.Now().Add(osdDuration)
	a.dirty = true
}

// flashPlayPause shows a large play or pause symbol fading out in the
// middle of the picture.
func (u *ui) flashPlayPause(playing bool) {
	u.flash = "pause"
	if playing {
		u.flash = "play"
	}
	u.flashAt = time.Now()
}

func (a *App) toast(msg string)      { a.ui.toasts.Show(toaster.Info, msg) }
func (a *App) toastError(msg string) { a.ui.toasts.Show(toaster.Error, msg) }
