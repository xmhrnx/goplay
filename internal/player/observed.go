package player

import (
	"encoding/json"
	"math"
	"strconv"

	"github.com/mehran9675/goplay/internal/mpv"
)

// observed lists the properties the player follows, the format each is
// delivered in, and how it updates State. Lists are observed without a
// format and re-read as JSON when they change.
var observed = []struct {
	name   string
	format mpv.Format
	apply  func(p *Player, ev *mpv.Event)
}{
	{"path", mpv.FormatString, func(p *Player, ev *mpv.Event) {
		p.State.Path = ev.String
		if t := p.titles[ev.String]; t != "" {
			p.State.Title = t
		}
	}},
	{"media-title", mpv.FormatString, func(p *Player, ev *mpv.Event) { p.State.Title = p.titleOf(p.State.Path, ev.String) }},
	{"idle-active", mpv.FormatFlag, func(p *Player, ev *mpv.Event) {
		p.State.Idle = ev.Flag
		if ev.Flag {
			p.State.Loading = false
		}
	}},
	{"duration", mpv.FormatDouble, func(p *Player, ev *mpv.Event) { p.State.Duration = ev.Double }},
	{"time-pos", mpv.FormatDouble, func(p *Player, ev *mpv.Event) { p.State.Position = ev.Double }},
	{"pause", mpv.FormatFlag, func(p *Player, ev *mpv.Event) { p.State.Paused = ev.Flag }},
	{"paused-for-cache", mpv.FormatFlag, func(p *Player, ev *mpv.Event) { p.State.Buffering = ev.Flag }},
	{"eof-reached", mpv.FormatFlag, func(p *Player, ev *mpv.Event) { p.State.EOF = ev.Flag }},
	{"seekable", mpv.FormatFlag, func(p *Player, ev *mpv.Event) { p.State.Seekable = ev.Flag }},
	{"demuxer-cache-duration", mpv.FormatDouble, func(p *Player, ev *mpv.Event) { p.State.CacheAhead = ev.Double }},
	{"demuxer-cache-state", mpv.FormatNone, func(p *Player, ev *mpv.Event) { p.readCacheRanges() }},
	{"volume", mpv.FormatDouble, func(p *Player, ev *mpv.Event) { p.State.Volume = ev.Double }},
	{"volume-max", mpv.FormatDouble, func(p *Player, ev *mpv.Event) { p.State.VolumeMax = ev.Double }},
	{"mute", mpv.FormatFlag, func(p *Player, ev *mpv.Event) { p.State.Muted = ev.Flag }},
	{"speed", mpv.FormatDouble, func(p *Player, ev *mpv.Event) { p.State.Speed = ev.Double }},
	{"track-list", mpv.FormatNone, func(p *Player, ev *mpv.Event) {
		p.State.Tracks = nil
		p.h.JSON("track-list", &p.State.Tracks)
		p.noteAudioTrack()
	}},
	{"audio-device-list", mpv.FormatNone, func(p *Player, ev *mpv.Event) { p.audioDevicesChanged() }},
	{"chapter-list", mpv.FormatNone, func(p *Player, ev *mpv.Event) {
		p.State.Chapters = nil
		p.h.JSON("chapter-list", &p.State.Chapters)
	}},
	{"chapter", mpv.FormatInt64, func(p *Player, ev *mpv.Event) { p.State.Chapter = intOr(ev, -1) }},
	{"edition-list", mpv.FormatNone, func(p *Player, ev *mpv.Event) {
		p.State.Editions = nil
		p.h.JSON("edition-list", &p.State.Editions)
	}},
	{"current-edition", mpv.FormatInt64, func(p *Player, ev *mpv.Event) { p.State.Edition = intOr(ev, 0) }},
	{"playlist", mpv.FormatNone, func(p *Player, ev *mpv.Event) {
		p.State.Playlist = nil
		p.h.JSON("playlist", &p.State.Playlist)
		for i, e := range p.State.Playlist {
			if e.Title == "" {
				p.State.Playlist[i].Title = p.titles[e.Filename]
			}
		}
	}},
	{"playlist-pos", mpv.FormatInt64, func(p *Player, ev *mpv.Event) { p.State.PlaylistPos = intOr(ev, -1) }},
	{"dwidth", mpv.FormatInt64, func(p *Player, ev *mpv.Event) { p.State.VideoW = intOr(ev, 0) }},
	{"dheight", mpv.FormatInt64, func(p *Player, ev *mpv.Event) { p.State.VideoH = intOr(ev, 0) }},
	{"ab-loop-a", mpv.FormatString, func(p *Player, ev *mpv.Event) { p.State.ABLoopA = parseTime(ev.String) }},
	{"ab-loop-b", mpv.FormatString, func(p *Player, ev *mpv.Event) { p.State.ABLoopB = parseTime(ev.String) }},
	{"loop-file", mpv.FormatString, func(p *Player, ev *mpv.Event) { p.State.LoopFile = isLooping(ev.String) }},
	{"loop-playlist", mpv.FormatString, func(p *Player, ev *mpv.Event) { p.State.LoopPlaylist = isLooping(ev.String) }},
	{"sub-delay", mpv.FormatDouble, func(p *Player, ev *mpv.Event) { p.State.SubDelay = ev.Double }},
	{"audio-delay", mpv.FormatDouble, func(p *Player, ev *mpv.Event) { p.State.AudioDelay = ev.Double }},
	{"hwdec-current", mpv.FormatString, func(p *Player, ev *mpv.Event) { p.State.Hwdec = ev.String }},
}

func intOr(ev *mpv.Event, def int) int {
	if ev.Format != mpv.FormatInt64 {
		return def
	}
	return int(ev.Int)
}

// parseTime parses an A-B loop point; "no" and unset give NaN.
func parseTime(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return math.NaN()
	}
	return f
}

func isLooping(s string) bool { return s != "" && s != "no" }

// readCacheRanges reads the buffered ranges from demuxer-cache-state.
func (p *Player) readCacheRanges() {
	var st struct {
		Ranges []struct {
			Start float64 `json:"start"`
			End   float64 `json:"end"`
		} `json:"seekable-ranges"`
	}
	p.State.CacheRanges = p.State.CacheRanges[:0]
	s, ok := p.h.String("demuxer-cache-state")
	if !ok || json.Unmarshal([]byte(s), &st) != nil {
		return
	}
	for _, r := range st.Ranges {
		p.State.CacheRanges = append(p.State.CacheRanges, [2]float64{r.Start, r.End})
	}
}
