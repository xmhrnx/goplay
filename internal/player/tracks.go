package player

import "strconv"

// Tracks.

// SelectTrack plays track id of kind; 0 turns the kind off.
func (p *Player) SelectTrack(k TrackKind, id int) {
	if k == Audio {
		p.audioTrack = id
	}
	p.h.SetString(fileLocal(trackProperty(k)), idOrNo(id))
}

// SelectSecondarySubtitle shows track id as a second subtitle line; 0
// turns it off.
func (p *Player) SelectSecondarySubtitle(id int) {
	p.h.SetString(fileLocal("secondary-sid"), idOrNo(id))
}

// fileLocal names option for the playing file only. A track picked by
// setting aid, sid or vid stays picked for the files after it, where the
// same number is another track or none: audio turned off once would stay
// off.
func fileLocal(option string) string { return "file-local-options/" + option }

// AddSubtitle loads an external subtitle file and shows it.
func (p *Player) AddSubtitle(path string) error {
	return p.h.Command("sub-add", path, "select")
}

// AddAudio loads an external audio track and plays it.
func (p *Player) AddAudio(path string) error {
	return p.h.Command("audio-add", path, "select")
}

// SetSubDelay sets the subtitle delay in seconds.
func (p *Player) SetSubDelay(d float64) { p.h.SetDouble("sub-delay", d) }

// SubStep shows the next (or previous) subtitle line by seeking to it.
func (p *Player) SubStep(d int) { p.h.Command("sub-seek", strconv.Itoa(d)) }

func trackProperty(k TrackKind) string {
	switch k {
	case Video:
		return "vid"
	case Audio:
		return "aid"
	}
	return "sid"
}
