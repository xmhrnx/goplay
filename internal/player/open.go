package player

import (
	"fmt"
	"strings"

	"github.com/mehran9675/goplay/internal/library"
)

// Opening.

// Open plays files and URLs, replacing the playlist, or appends them.
func (p *Player) Open(items []string, appendToPlaylist bool) error {
	if len(items) == 0 {
		return nil
	}
	mode := "replace"
	if appendToPlaylist {
		mode = "append-play"
	} else if e, _ := p.h.String("options/edition"); e != "auto" {
		// A title picked with SetEdition was for the last disc.
		p.h.SetString("options/edition", "auto")
	}
	for i, it := range items {
		if err := p.h.Command("loadfile", it, mode); err != nil {
			return fmt.Errorf("open %s: %w", it, err)
		}
		if i == 0 {
			mode = "append"
			if appendToPlaylist {
				mode = "append-play"
			}
		}
	}
	return nil
}

// Name gives the item at url a title, shown for it instead of the
// address mpv would show: a media server's files are addresses with an
// ID in them, and the server lists their names.
func (p *Player) Name(url, title string) {
	if title != "" {
		p.titles[url] = title
	}
}

// titleOf is the title to show for path: the name given with Name, or
// mpv's media-title. A disc without a title of its own is named after
// its folder, image or drive; mpv would show the end of its address.
func (p *Player) titleOf(path, mediaTitle string) string {
	if t := p.titles[path]; t != "" {
		return t
	}
	if library.IsDiscURL(path) && (mediaTitle == "" || strings.Contains(path, mediaTitle)) {
		return library.DiscName(path)
	}
	return mediaTitle
}

// Stop ends playback and clears the playlist.
func (p *Player) Stop() { p.h.Command("stop") }
