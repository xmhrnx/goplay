package player

import "strconv"

// Playlist.

// PlaylistNext and PlaylistPrev move through the playlist.
func (p *Player) PlaylistNext() { p.h.Command("playlist-next") }
func (p *Player) PlaylistPrev() { p.h.Command("playlist-prev") }

// PlaylistPlay plays entry i.
func (p *Player) PlaylistPlay(i int) { p.h.Command("playlist-play-index", strconv.Itoa(i)) }

// PlaylistRemove removes entry i.
func (p *Player) PlaylistRemove(i int) { p.h.Command("playlist-remove", strconv.Itoa(i)) }

// PlaylistMove moves entry from to before entry to.
func (p *Player) PlaylistMove(from, to int) {
	p.h.Command("playlist-move", strconv.Itoa(from), strconv.Itoa(to))
}

// PlaylistShuffle shuffles the playlist.
func (p *Player) PlaylistShuffle() { p.h.Command("playlist-shuffle") }

// PlaylistClear removes every entry but the playing one.
func (p *Player) PlaylistClear() { p.h.Command("playlist-clear") }
