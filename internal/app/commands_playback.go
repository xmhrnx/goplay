package app

import (
	"math"
	"strconv"

	"github.com/mehran9675/goplay/internal/player"
	"github.com/mehran9675/goplay/internal/settings"
)

func (a *App) seekStep(large bool) float64 {
	if large {
		return a.settings.Float(settings.KeySeekStepLarge)
	}
	return a.settings.Float(settings.KeySeekStep)
}

// Commands that work the same whether playing here or casting.

func (a *App) togglePause() {
	if a.casting() {
		a.castTogglePause()
		return
	}
	st := &a.player.State
	a.player.TogglePause()
	a.ui.flashPlayPause(!st.Paused || st.EOF)
}

func (a *App) seekTo(t float64) {
	if a.casting() {
		a.castSeek(t)
		return
	}
	a.player.Seek(t, false)
}

func (a *App) seekBy(d float64) {
	if a.casting() {
		a.castSeek(a.cast.status.Position + d)
		return
	}
	a.player.SeekBy(d)
	sign := "+"
	if d < 0 {
		sign = "−"
	}
	a.osd(sign + player.FormatTime(math.Abs(d)))
}

func (a *App) previous() {
	st := &a.player.State
	// Like a CD player: back to the start first, then the previous file.
	if st.Position > 5 || st.PlaylistPos <= 0 {
		a.seekTo(0)
		return
	}
	a.player.PlaylistPrev()
}

func (a *App) setSpeed(s float64) {
	a.player.SetSpeed(s)
	a.osd(tr("Speed {speed}", "speed", speedLabel(s)))
}

func (a *App) stepSpeed(dir int) {
	cur := a.player.State.Speed
	next := cur
	if dir > 0 {
		for _, s := range speedSteps {
			if s > cur+1e-6 {
				next = s
				break
			}
		}
	} else {
		for i := len(speedSteps) - 1; i >= 0; i-- {
			if speedSteps[i] < cur-1e-6 {
				next = speedSteps[i]
				break
			}
		}
	}
	a.setSpeed(next)
}

func speedLabel(s float64) string {
	if s == 1 {
		return "1×"
	}
	return strconv.FormatFloat(s, 'f', -1, 64) + "×"
}

func (a *App) cycleABLoop() {
	st := &a.player.State
	a.player.CycleABLoop()
	switch {
	case math.IsNaN(st.ABLoopA):
		a.osd(tr("Section start set — press again to set the end"))
	case math.IsNaN(st.ABLoopB):
		a.osd(tr("Repeating section"))
	default:
		a.osd(tr("Section repeat off"))
	}
}
