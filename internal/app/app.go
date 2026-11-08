// Package app is goplay itself: it wires the player (mpv), the window
// and the Giode UI together, and holds the views.
//
// Files: app.go (the window and frames), commands.go (every command),
// media.go (opening files and URLs), stereo.go (3D and 360°), casting.go
// (Chromecast and DLNA), theme.go and icons.go (look), and view_*.go
// (the UI).
//
// The window is a giode.App: mpv draws each video frame in its
// underlay, straight into the window's OpenGL framebuffer, and the UI is
// drawn over it.
//
// Threading: everything here runs on the UI goroutine, except the
// underlay (drawVideo), which runs on the window's render thread and
// only touches the render context. mpv's threads and background
// goroutines hand work to the UI goroutine with Post.
package app

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"sync/atomic"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/elements"

	"github.com/mehran9675/goplay/internal/actions"
	"github.com/mehran9675/goplay/internal/library"
	"github.com/mehran9675/goplay/internal/mpv"
	"github.com/mehran9675/goplay/internal/player"
	"github.com/mehran9675/goplay/internal/settings"
)

// App is the running player.
type App struct {
	g        *giode.App
	player   *player.Player
	settings *settings.Store
	history  *library.History
	actions  *actions.Registry
	ui       *ui

	configDir string

	// The render context lives on the render thread (see drawVideo).
	rend       *mpv.RenderContext
	rendFailed bool
	// showVideo tells the render thread whether to draw mpv's picture:
	// not while idle or casting, when the UI paints its own background.
	showVideo atomic.Bool

	// mpvPending is set while a Post to process mpv's events is queued.
	mpvPending atomic.Bool

	// dirty asks for another frame once this one is done.
	dirty bool

	// The window: its title as last set and its size in pixels.
	title   string
	winSize image.Point

	// starting runs once the video output is ready (or failed): the files
	// and commands from the command line. mpv turns video off for a file
	// that starts before it.
	starting []func()
	ready    bool

	// pausedHidden is set when minimizing paused playback, to resume it
	// on the way back.
	pausedHidden bool

	logMu sync.Mutex
	log   []string // mpv log lines, newest last

	stereo stereoState
	cast   castState

	// audioChoice is the audio track the viewer last picked, played in
	// the files after it until GoPlay closes (see pickTrack).
	audioChoice atomic.Pointer[player.TrackChoice]

	// lastPath and lastPos remember the file being played, for the
	// recent list when it ends.
	lastPath  string
	lastTitle string
	// resumeSaved is set once Quit has saved where playback was, so
	// shutdown doesn't save it again, paused.
	resumeSaved bool

	// startupCommands run once the first file has loaded (--run).
	startupCommands []string
}

// Version is goplay's version, set at build time with -ldflags.
var Version = "dev"

// New sets up the player and window. files are opened, then commands
// (action IDs) run, once the window shows. icon is the window icon, as
// PNG.
func New(files, commands []string, icon []byte) (*App, error) {
	// Most of a video player's memory is mpv's; keep the Go side lean.
	debug.SetGCPercent(50)

	a := &App{actions: actions.NewRegistry()}
	dir, err := settings.ConfigDir()
	if err == nil {
		a.configDir = dir
	}
	a.loadPrefs()

	a.g = giode.New(giode.Config{
		Title:       "GoPlay",
		Locale:      uiLocale(a.settings.Get(settings.KeyLanguage)),
		Intl:        intlOptions,
		Size:        image.Pt(960, 600),
		MinSize:     image.Pt(360, 240),
		Resizable:   true,
		Frameless:   true,
		Center:      true,
		AlwaysOnTop: a.settings.Bool(settings.KeyAlwaysOnTop),
		Background:  colBackground,
		Icon:        icon,
		Renderer:    giode.RendererOpenGL,
		Underlay:    a.drawVideo,
		// mpv's render context must be freed while the GL context is
		// current.
		UnderlayRelease: a.releaseVideo,
		OnEvent:         a.windowEvent,
		OnKey:           a.onKey,
		OnFileDrop:      a.dropped,
		OnOpenURL:       a.openedURL,
		OnContextAction: func(_ string, paths []string) {
			a.whenReady(func() { a.openItems(paths, false) })
		},
		OnStartup: func() {
			if runtime.GOOS == "linux" {
				// Linux file managers find "Play with GoPlay" in the
				// user's home; the other systems' installers add it.
				a.g.InstallContextActions()
			}
		},
		OnWindowStateChange: a.windowStateChanged,
		OnShutdown:          a.shutdown,
	})

	if err := a.startPlayer(); err != nil {
		return nil, err
	}

	setupTheme()
	a.registerCommands()
	a.applyKeyBindings()
	a.ui = newUI(a)

	if len(files) > 0 {
		a.starting = append(a.starting, func() { a.openItems(files, false) })
	}
	switch {
	case len(commands) > 0 && len(files) > 0:
		// Commands that act on the file wait for it to load.
		a.startupCommands = commands
	case len(commands) > 0:
		a.starting = append(a.starting, func() { a.runCommands(commands) })
	}
	return a, nil
}

func (a *App) loadPrefs() {
	path, histPath := "", ""
	if a.configDir != "" {
		path = filepath.Join(a.configDir, "settings.json")
		histPath = filepath.Join(a.configDir, "recent.json")
	}
	s, err := settings.Load(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "GoPlay: settings:", err)
	}
	a.settings = s
	h, err := library.LoadHistory(histPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "GoPlay: recent files:", err)
	}
	a.history = h
}

// startPlayer creates mpv with the user's settings. Its render context
// is made on the render thread, on the first frame (drawVideo).
func (a *App) startPlayer() error {
	var opts []player.Option
	for _, o := range a.settings.MPVOptions() {
		opts = append(opts, player.Option{Name: o.Name, Value: o.Value})
	}
	opts = append(opts, player.Option{Name: "alang", Value: a.audioLanguages()})
	if a.settings.Bool(settings.KeyRememberVolume) && a.settings.Get(settings.KeyVolume) != "" {
		opts = append(opts, player.Option{Name: "volume", Value: a.settings.Get(settings.KeyVolume)})
	}
	p, err := player.New(player.Config{
		Options:    opts,
		ConfigDir:  a.configDir,
		LoadConfig: a.settings.Bool(settings.KeyMPVConf),
		LogLevel:   a.settings.Get(settings.KeyLogLevel),
		Wakeup: func() {
			if !a.mpvPending.Swap(true) {
				a.g.Post(a.processPlayer)
			}
		},
	})
	var optErr *player.OptionError
	switch {
	case p == nil:
		return fmt.Errorf("mpv: %w", err)
	case err != nil && asOptionError(err, &optErr):
		a.starting = append(a.starting, func() { a.toastError(optErr.Error()) })
	}
	a.player = p
	p.OnLog = a.appendLog
	p.OnFileLoaded = a.fileLoaded
	p.OnEndFile = a.fileEnded
	return nil
}

func asOptionError(err error, target **player.OptionError) bool {
	oe, ok := err.(*player.OptionError)
	if ok {
		*target = oe
	}
	return ok
}

// Run shows the window and runs until it closes; the process ends
// with it.
func (a *App) Run() error {
	return a.g.Run(func() giode.Element { return elements.Raw(a.frame) })
}
