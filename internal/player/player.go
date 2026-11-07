// Package player drives an mpv core for goplay: it configures mpv,
// mirrors the properties the UI shows into State, and offers typed
// commands for everything the UI does. It knows nothing about windows or
// widgets; the app renders the video (mpv.RenderContext) and draws the
// UI over it.
//
// Threading: Process and every command are called on the UI goroutine.
// mpv's own threads only call the wakeup function given to New, which
// must just wake that goroutine.
package player

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/mehran9675/goplay/internal/mpv"
)

// Option is an mpv option name and value.
type Option struct{ Name, Value string }

// Config configures a Player.
type Config struct {
	// Options are mpv options applied after goplay's base options,
	// typically from the user's settings.
	Options []Option
	// ConfigDir is where mpv reads mpv.conf, input.conf and scripts, and
	// writes resume data. Empty disables config files.
	ConfigDir string
	// LoadConfig loads mpv.conf, input.conf and scripts from ConfigDir.
	LoadConfig bool
	// Wakeup is called from mpv's threads when events are waiting. It
	// must only wake the UI goroutine, which then calls Process.
	Wakeup func()
	// LogLevel is mpv's log level for Log ("warn" if empty).
	LogLevel string
}

// Player is an mpv core with the state the UI shows.
type Player struct {
	h     *mpv.Handle
	State State

	// OnFileLoaded runs when a file starts playing; OnEndFile when one
	// ends, with the reason (mpv.EndEOF, mpv.EndError, ...) and error.
	OnFileLoaded func()
	OnEndFile    func(reason int, err error)
	// OnLog receives mpv's log messages at the configured level.
	OnLog func(level, prefix, text string)

	filters map[string]string // labeled video filters goplay manages
	// hwdecSaved is the user's hardware decoding mode while goplay's
	// filters need decoded frames in system memory ("" when not
	// switched).
	hwdecSaved string
	// audioDevice is the output device the user chose ("auto" for the
	// default); devices is mpv's device list as last seen, "" before mpv
	// sent it. audioTrack is the audio track playing, or last playing before mpv
	// gave up on its output; 0 when the user turned audio off.
	audioDevice string
	devices     string
	audioTrack  int
	// titles names items opened by address, which mpv would show as
	// the address (see Name).
	titles map[string]string

	wake chan struct{} // tests only: woken with the core
}

// OptionError lists options mpv rejected at startup. The player still
// works; those options have their defaults.
type OptionError struct{ Options []string }

// baseOptions are what goplay itself needs from mpv, set before the
// user's settings (which come after and can override most of these).
var baseOptions = []Option{
	// Video goes through the render API, drawn by the app.
	{"vo", "libmpv"},
	// The UI is goplay's own: no on-screen controller or OSD bar, and
	// no keyboard handling inside mpv (keys are forwarded on purpose).
	{"osc", "no"},
	{"osd-bar", "no"},
	{"input-default-bindings", "no"},
	{"input-vo-keyboard", "no"},
	// Stay on the last frame at the end of the last file instead of
	// going idle, so the picture doesn't vanish and Play replays.
	{"keep-open", "yes"},
	{"idle", "yes"},
	{"screenshot-template", "GoPlay-%F-%wH.%wM.%wS"},
	// Built-in scripts that goplay's own UI replaces. Older mpv
	// versions lack some of these options; they're skipped there.
	{"load-console", "no"},
	{"load-select", "no"},
	{"load-positioning", "no"},
	{"load-commands", "no"},
	{"load-context-menu", "no"},
}

// New creates and starts the mpv core.
func New(cfg Config) (*Player, error) {
	h, err := mpv.Create()
	if err != nil {
		return nil, err
	}
	p := &Player{h: h, filters: map[string]string{}, titles: map[string]string{}}
	p.State.ABLoopA, p.State.ABLoopB = math.NaN(), math.NaN()
	p.State.Idle = true
	p.State.Speed = 1
	p.State.Chapter = -1
	p.State.PlaylistPos = -1

	base := append(append([]Option(nil), baseOptions...), platformOptions...)
	if pics := picturesDir(); pics != "" {
		base = append(base, Option{"screenshot-dir", pics})
	}
	if cfg.ConfigDir != "" {
		base = append(base,
			Option{"config-dir", cfg.ConfigDir},
			Option{"watch-later-dir", filepath.Join(cfg.ConfigDir, "watch_later")},
		)
	}
	base = append(base, Option{"config", yesNo(cfg.LoadConfig && cfg.ConfigDir != "")})
	for _, o := range base {
		if err := h.SetOption(o.Name, o.Value); err != nil && o.Name == "vo" {
			h.Destroy()
			return nil, fmt.Errorf("mpv: %s=%s: %w", o.Name, o.Value, err)
		}
	}
	p.audioDevice = "auto"
	var bad []string
	for _, o := range cfg.Options {
		if o.Name == "audio-device" {
			p.audioDevice = o.Value
		}
		if err := h.SetOption(o.Name, o.Value); err != nil {
			// A bad saved value must not stop the player from starting:
			// mpv keeps its default and the caller is told.
			bad = append(bad, fmt.Sprintf("%s=%s: %v", o.Name, o.Value, err))
		}
	}
	if err := h.Initialize(); err != nil {
		h.Destroy()
		return nil, err
	}
	level := cfg.LogLevel
	if level == "" {
		level = "warn"
	}
	h.RequestLog(level)
	for i, o := range observed {
		h.Observe(uint64(i+1), o.name, o.format)
	}
	if cfg.Wakeup != nil {
		h.SetWakeupCallback(cfg.Wakeup)
	}
	if len(bad) > 0 {
		return p, &OptionError{Options: bad}
	}
	return p, nil
}

func (e *OptionError) Error() string {
	return "mpv rejected some settings: " + strings.Join(e.Options, "; ")
}

// Handle is the underlying mpv core, for rendering and for commands the
// Player doesn't wrap.
func (p *Player) Handle() *mpv.Handle { return p.h }

// Process handles all waiting mpv events, updating State and running
// the callbacks. It reports whether anything changed.
func (p *Player) Process() bool {
	changed := false
	for {
		ev := p.h.Wait(0)
		switch ev.ID {
		case mpv.EventNone:
			return changed
		case mpv.EventPropertyChange:
			if i := int(ev.Reply) - 1; i >= 0 && i < len(observed) {
				observed[i].apply(p, &ev)
				changed = true
			}
		case mpv.EventStartFile:
			p.State.Loading = true
			p.audioTrack = 0
			changed = true
		case mpv.EventFileLoaded:
			p.State.Loading = false
			changed = true
			// The property updates for the new file can still be queued
			// behind this event: read what the callback needs now.
			p.State.Path, _ = p.h.String("path")
			mt, _ := p.h.String("media-title")
			p.State.Title = p.titleOf(p.State.Path, mt)
			p.State.Duration, _ = p.h.Double("duration")
			p.State.Tracks = nil
			p.h.JSON("track-list", &p.State.Tracks)
			p.noteAudioTrack()
			if p.OnFileLoaded != nil {
				p.OnFileLoaded()
			}
		case mpv.EventEndFile:
			changed = true
			if ev.EndReason != mpv.EndRedirect {
				p.State.Loading = false
			}
			if p.OnEndFile != nil {
				p.OnEndFile(ev.EndReason, ev.Error)
			}
		case mpv.EventLogMessage:
			if p.OnLog != nil {
				p.OnLog(ev.LogLevel, ev.LogPrefix, ev.LogText)
			}
		case mpv.EventShutdown:
			return true
		}
	}
}

// Destroy stops mpv. Free the render context first.
func (p *Player) Destroy() { p.h.Destroy() }
