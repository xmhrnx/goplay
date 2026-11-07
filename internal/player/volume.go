package player

import "math"

// Audio.

// SetVolume sets the volume in percent (up to volume-max).
func (p *Player) SetVolume(v float64) {
	max := p.State.VolumeMax
	if max <= 0 {
		max = 100
	}
	p.h.SetDouble("volume", math.Max(0, math.Min(max, v)))
}

// SetMute mutes or unmutes.
func (p *Player) SetMute(m bool) { p.h.SetFlag("mute", m) }

// SetAudioDelay sets the audio delay in seconds (positive delays it).
func (p *Player) SetAudioDelay(d float64) { p.h.SetDouble("audio-delay", d) }
