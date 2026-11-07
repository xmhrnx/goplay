package library

import (
	"path/filepath"
	"time"
)

// Recent is one recently played file or URL.
type Recent struct {
	Path     string    `json:"path"`
	Title    string    `json:"title,omitempty"`
	Position float64   `json:"position,omitempty"`
	Duration float64   `json:"duration,omitempty"`
	Played   time.Time `json:"played"`
}

// Progress is how far through it playback got, from 0 to 1.
func (r Recent) Progress() float64 {
	if r.Duration <= 0 {
		return 0
	}
	return min(1, max(0, r.Position/r.Duration))
}

// Name is the title, or the file name.
func (r Recent) Name() string {
	if r.Title != "" {
		return r.Title
	}
	if IsDiscURL(r.Path) {
		return DiscName(r.Path)
	}
	if IsURL(r.Path) {
		return r.Path
	}
	return filepath.Base(r.Path)
}
