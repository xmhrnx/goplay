package player

// Generic access, for settings and power users.

// Set sets any mpv property or option at runtime.
func (p *Player) Set(name, value string) error { return p.h.SetString(name, value) }

// Get reads any mpv property as a string.
func (p *Player) Get(name string) (string, bool) { return p.h.String(name) }

// Command runs any mpv command.
func (p *Player) Command(args ...string) error { return p.h.Command(args...) }

// Keypress forwards a key to mpv, so bindings in input.conf and scripts
// work. name uses mpv's key names ("ctrl+a", "F1", "KP_ENTER").
func (p *Player) Keypress(name string) { p.h.Command("keypress", name) }

// SaveResumePosition stores the current file's position so it resumes
// there next time (mpv saves on quit by itself).
func (p *Player) SaveResumePosition() {
	if !p.State.Idle && !p.State.EOF {
		p.h.Command("write-watch-later-config")
	}
}
