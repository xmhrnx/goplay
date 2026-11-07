package app

import (
	"fmt"
	"math"
	"strconv"

	"github.com/mehran9675/goplay/internal/player"
)

func (a *App) addVolume(d float64) {
	st := &a.player.State
	if a.casting() {
		v := math.Max(0, math.Min(1, a.cast.status.Volume+d/100))
		a.castVolume(v)
		a.osd(tr("Volume {percent}%", "percent", int(math.Round(v*100))))
		return
	}
	v := st.Volume + d
	a.player.SetVolume(v)
	if st.Muted && d > 0 {
		a.player.SetMute(false)
	}
	max := st.VolumeMax
	if max <= 0 {
		max = 100
	}
	a.osdLevel(tr("Volume {percent}%", "percent", int(math.Round(math.Max(0, math.Min(max, v))))), math.Max(0, math.Min(max, v))/max)
}

func (a *App) setVolume(v float64) {
	a.player.SetVolume(v)
	if a.player.State.Muted && v > 0 {
		a.player.SetMute(false)
	}
}

func (a *App) toggleMute() {
	m := !a.player.State.Muted
	a.player.SetMute(m)
	a.osd(onOff(tr("Mute"), m))
}

func (a *App) addAudioDelay(d float64) {
	v := a.player.State.AudioDelay + d
	a.player.SetAudioDelay(v)
	a.osd(tr("Audio delay {seconds} s", "seconds", fmt.Sprintf("%+.1f", v)))
}

func (a *App) addSubDelay(d float64) {
	v := a.player.State.SubDelay + d
	a.player.SetSubDelay(v)
	a.osd(tr("Subtitle delay {seconds} s", "seconds", fmt.Sprintf("%+.1f", v)))
}

func (a *App) scaleSubs(d float64) {
	cur, ok := a.player.Handle().Double("sub-scale")
	if !ok {
		cur = 1
	}
	v := math.Max(0.3, math.Min(4, cur+d))
	a.player.Set("sub-scale", strconv.FormatFloat(v, 'f', 2, 64))
	a.osd(tr("Subtitle size {percent}%", "percent", int(math.Round(v*100))))
}

func (a *App) toggleSubtitles() {
	st := &a.player.State
	if _, on := st.SelectedTrack(player.Subtitle); on {
		a.pickTrack(player.Subtitle, 0)
		a.osd(tr("Subtitles off"))
		return
	}
	subs := st.TracksOf(player.Subtitle)
	if len(subs) > 0 {
		a.pickTrack(player.Subtitle, subs[0].ID)
		a.osd(tr("Subtitles: {track}", "track", subs[0].Name()))
	}
}

// cycleTrack moves to the next track of a kind (subtitles cycle through
// "off" too).
func (a *App) cycleTrack(k player.TrackKind) {
	st := &a.player.State
	tracks := st.TracksOf(k)
	if len(tracks) == 0 {
		return
	}
	cur, _ := st.SelectedTrack(k)
	idx := -1
	for i, t := range tracks {
		if t.ID == cur.ID {
			idx = i
		}
	}
	next := idx + 1
	if next >= len(tracks) {
		if k == player.Subtitle {
			a.pickTrack(k, 0)
			a.osd(tr("Subtitles off"))
			return
		}
		next = 0
	}
	a.pickTrack(k, tracks[next].ID)
	label := map[player.TrackKind]string{player.Video: tr("Video"), player.Audio: tr("Audio"), player.Subtitle: tr("Subtitles")}[k]
	a.osd(tr("{kind}: {track}", "kind", label, "track", tracks[next].Name()))
}
