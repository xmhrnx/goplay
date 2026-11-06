package app

import (
	"encoding/json"

	"github.com/mehran9675/goplay/internal/player"
	"github.com/mehran9675/goplay/internal/settings"
)

// pickTrack plays track id of kind k (0 turns it off) because the viewer
// chose it, and remembers the choice for the files after: the audio
// track until GoPlay closes, the subtitle track for good.
func (a *App) pickTrack(k player.TrackKind, id int) {
	a.player.SelectTrack(k, id)
	c := player.ChoiceOf(a.player.State.Tracks, k, id)
	switch k {
	case player.Audio:
		a.audioChoice.Store(&c)
	case player.Subtitle:
		data, err := json.Marshal(c)
		if err != nil {
			return
		}
		a.settings.Set(settings.KeySubtitleChoice, string(data))
	}
}

// applyTrackChoices picks the remembered audio and subtitle tracks in
// the file that just loaded, where it has them; otherwise mpv's own pick
// (the preferred languages) stays.
func (a *App) applyTrackChoices() {
	tracks := a.player.State.Tracks
	if c := a.audioChoice.Load(); c != nil {
		if id, ok := c.Match(tracks, player.Audio); ok {
			a.player.SelectTrack(player.Audio, id)
		}
	}
	var sub player.TrackChoice
	if s := a.settings.Get(settings.KeySubtitleChoice); s != "" && json.Unmarshal([]byte(s), &sub) == nil {
		if id, ok := sub.Match(tracks, player.Subtitle); ok {
			a.player.SelectTrack(player.Subtitle, id)
		}
	}
}
