package player

import (
	"strconv"
	"strings"

	"github.com/xmhrnx/giode/i18n"
)

// TrackKind is a track type, named as mpv names it.
type TrackKind string

// Track is one video, audio or subtitle track of the current file.
type Track struct {
	ID               int       `json:"id"`
	Kind             TrackKind `json:"type"`
	Title            string    `json:"title"`
	Lang             string    `json:"lang"`
	Codec            string    `json:"codec"`
	Selected         bool      `json:"selected"`
	Default          bool      `json:"default"`
	Forced           bool      `json:"forced"`
	External         bool      `json:"external"`
	ExternalFilename string    `json:"external-filename"`
	AlbumArt         bool      `json:"albumart"`
	Width            int       `json:"demux-w"`
	Height           int       `json:"demux-h"`
	FPS              float64   `json:"demux-fps"`
	Channels         int       `json:"demux-channel-count"`
	SampleRate       int       `json:"demux-samplerate"`
	// MainSelection is mpv's index of the selection slot this track is in:
	// 0 for the primary track, 1 for the secondary subtitle.
	MainSelection int `json:"main-selection"`
}

// Chapter is a chapter of the current file.
type Chapter struct {
	Title string  `json:"title"`
	Time  float64 `json:"time"`
}

// Edition is one of the file's editions: a Matroska file's alternative
// cuts, or a DVD's or Blu-ray's titles.
type Edition struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// PlaylistEntry is one entry of mpv's playlist.
type PlaylistEntry struct {
	Filename string `json:"filename"`
	Title    string `json:"title"`
	Current  bool   `json:"current"`
}

// State mirrors the player's observed properties. It is updated on the
// UI goroutine by Player.Process and read freely there.
type State struct {
	Path  string
	Title string
	// Idle is true with nothing loaded.
	Idle bool
	// Loading is true from opening a file until it starts playing.
	Loading bool

	Duration, Position float64
	Paused, Buffering  bool
	EOF, Seekable      bool
	// CacheAhead is how many seconds are buffered ahead of the position.
	CacheAhead float64
	// CacheRanges are the buffered, seekable time ranges (streams).
	CacheRanges [][2]float64

	Volume, VolumeMax float64 // percent
	Muted             bool
	Speed             float64

	Tracks       []Track
	Chapters     []Chapter
	Chapter      int // -1 before the first chapter or without chapters
	Editions     []Edition
	Edition      int // the playing edition's ID
	Playlist     []PlaylistEntry
	PlaylistPos  int
	VideoW       int
	VideoH       int
	ABLoopA      float64 // NaN when unset
	ABLoopB      float64
	LoopFile     bool
	LoopPlaylist bool
	SubDelay     float64
	AudioDelay   float64
	Hwdec        string
}

const (
	Video    TrackKind = "video"
	Audio    TrackKind = "audio"
	Subtitle TrackKind = "sub"
)

// Name is a readable label: the title (or file name of an external
// track), the codec and size for video, and the language.
func (t Track) Name() string {
	name := t.Title
	if name == "" && t.ExternalFilename != "" {
		name = baseName(t.ExternalFilename)
	}
	if name == "" {
		switch {
		case t.Kind == Video && t.Width > 0:
			name = strings.ToUpper(t.Codec) + " " + strconv.Itoa(t.Width) + "×" + strconv.Itoa(t.Height)
		case t.Kind == Audio && t.Channels > 0:
			name = strings.ToUpper(t.Codec) + " " + channelName(t.Channels)
		default:
			name = i18n.Intl("Track {number}", "number", t.ID)
		}
	}
	if t.Lang != "" {
		name += " · " + LanguageName(t.Lang)
	}
	if t.Forced {
		name += " " + i18n.Intl("(forced)")
	}
	return name
}

func channelName(n int) string {
	switch n {
	case 1:
		return i18n.Intl("mono")
	case 2:
		return i18n.Intl("stereo")
	case 6:
		return "5.1"
	case 8:
		return "7.1"
	}
	return strconv.Itoa(n) + "ch"
}

func baseName(p string) string {
	if i := strings.LastIndexAny(p, `/\`); i >= 0 {
		return p[i+1:]
	}
	return p
}

// Name is the entry's title, or its file name.
func (e PlaylistEntry) Name() string {
	if e.Title != "" {
		return e.Title
	}
	return baseName(strings.TrimRight(e.Filename, `/\`))
}

// HasVideo reports whether the file has a (non-cover-art) video track
// playing.
func (s *State) HasVideo() bool {
	for _, t := range s.Tracks {
		if t.Kind == Video && t.Selected && !t.AlbumArt {
			return true
		}
	}
	return false
}

// TracksOf returns the tracks of one kind.
func (s *State) TracksOf(k TrackKind) []Track {
	var out []Track
	for _, t := range s.Tracks {
		if t.Kind == k {
			out = append(out, t)
		}
	}
	return out
}

// SelectedTrack returns the primary selected track of kind k, if any.
func (s *State) SelectedTrack(k TrackKind) (Track, bool) {
	for _, t := range s.Tracks {
		if t.Kind == k && t.Selected && t.MainSelection == 0 {
			return t, true
		}
	}
	return Track{}, false
}

// SecondarySubtitle returns the secondary subtitle track, if any.
func (s *State) SecondarySubtitle() (Track, bool) {
	for _, t := range s.Tracks {
		if t.Kind == Subtitle && t.Selected && t.MainSelection == 1 {
			return t, true
		}
	}
	return Track{}, false
}

// ChapterAt returns the index of the chapter containing t, or -1.
func (s *State) ChapterAt(t float64) int {
	idx := -1
	for i, c := range s.Chapters {
		if c.Time <= t {
			idx = i
		}
	}
	return idx
}
