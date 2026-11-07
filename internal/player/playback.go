package player

import (
	"math"
	"strconv"
)

// Playback.

// TogglePause pauses or resumes; at the end of the file it replays.
func (p *Player) TogglePause() {
	if p.State.EOF && p.State.Paused {
		p.h.Command("seek", "0", "absolute")
		p.h.SetFlag("pause", false)
		return
	}
	p.h.Command("cycle", "pause")
}

// SetPause pauses or resumes.
func (p *Player) SetPause(paused bool) { p.h.SetFlag("pause", paused) }

// Seek jumps to t seconds; exact seeks land on the exact frame (slower).
func (p *Player) Seek(t float64, exact bool) {
	mode := "absolute+keyframes"
	if exact {
		mode = "absolute+exact"
	}
	p.h.Command("seek", ftoa(t), mode)
}

// SeekBy moves by d seconds.
func (p *Player) SeekBy(d float64) { p.h.Command("seek", ftoa(d), "relative") }

// SeekFraction jumps to a fraction [0, 1] of the duration.
func (p *Player) SeekFraction(f float64, exact bool) {
	if p.State.Duration > 0 {
		p.Seek(math.Max(0, math.Min(1, f))*p.State.Duration, exact)
	}
}

// FrameStep advances one frame, or goes back one; both pause.
func (p *Player) FrameStep(back bool) {
	if back {
		p.h.Command("frame-back-step")
	} else {
		p.h.Command("frame-step")
	}
}

// SetSpeed sets the playback speed (1 is normal).
func (p *Player) SetSpeed(s float64) { p.h.SetDouble("speed", math.Max(0.01, math.Min(100, s))) }

// Chapters.

// ChapterBy moves d chapters (negative goes back).
func (p *Player) ChapterBy(d int) { p.h.Command("add", "chapter", strconv.Itoa(d)) }

// SetChapter jumps to chapter i.
func (p *Player) SetChapter(i int) { p.h.SetString("chapter", strconv.Itoa(i)) }

// SetEdition plays the edition (a disc's title) with the given ID,
// reopening the file. It stays chosen until something else is opened
// (unlike a track, an edition can't be set for one file: reopening ends
// the file, which would undo it).
func (p *Player) SetEdition(id int) { p.h.SetString("edition", strconv.Itoa(id)) }

// A-B loop and repeat.

// CycleABLoop sets point A, then B, then clears the loop.
func (p *Player) CycleABLoop() { p.h.Command("ab-loop") }

// ClearABLoop removes the A-B loop.
func (p *Player) ClearABLoop() {
	p.h.SetString("ab-loop-a", "no")
	p.h.SetString("ab-loop-b", "no")
}

// SetLoopFile repeats the current file.
func (p *Player) SetLoopFile(on bool) { p.h.SetString("loop-file", infOrNo(on)) }

// SetLoopPlaylist repeats the playlist.
func (p *Player) SetLoopPlaylist(on bool) { p.h.SetString("loop-playlist", infOrNo(on)) }
