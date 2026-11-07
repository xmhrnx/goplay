package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mehran9675/goplay/internal/library"
	"github.com/mehran9675/goplay/internal/settings"
)

// openItems plays files, folders and URLs, or adds them to the playlist.
// A single file brings the rest of its folder along when that setting is
// on, starting at the file itself.
func (a *App) openItems(items []string, appendToPlaylist bool) {
	if len(items) == 0 {
		return
	}
	if a.settings.Bool("resume") {
		a.player.SaveResumePosition()
	}
	list, start := library.Expand(items, !appendToPlaylist && a.settings.Bool(settings.KeyQueueFolder))
	if len(list) == 0 {
		a.toastError(tr("There's nothing to play there."))
		return
	}
	if err := a.player.Open(list, appendToPlaylist); err != nil {
		a.toastError(err.Error())
		return
	}
	if appendToPlaylist {
		a.toast(tr("{count} files added to the playlist", "count", len(list)))
		return
	}
	if start > 0 {
		a.player.PlaylistPlay(start)
	}
	if a.casting() {
		a.Post(a.castCurrent)
		return
	}
	// mpv keeps pause across files, and a finished file is left paused on
	// its last frame: what's opened plays.
	a.player.SetPause(false)
}

// dropped handles files dropped on the window: subtitle files join the
// playing video; anything else plays.
func (a *App) dropped(paths []string) {
	var media []string
	for _, p := range paths {
		if library.IsSubtitle(p) && !a.player.State.Idle {
			a.addSubtitle(p)
			continue
		}
		media = append(media, p)
	}
	a.openItems(media, false)
}

// openClipboard plays a link or path from the clipboard.
func (a *App) openClipboard() {
	a.g.ReadClipboard(a.openText)
}

// openText plays the links or paths in text, one per line.
func (a *App) openText(text string) {
	text = strings.TrimSpace(text)
	if text == "" {
		a.toast(tr("The clipboard is empty"))
		return
	}
	var items []string
	for _, line := range strings.Split(text, "\n") {
		if line = strings.Trim(strings.TrimSpace(line), `"`); line != "" {
			items = append(items, line)
		}
	}
	for _, it := range items {
		if _, err := os.Stat(it); err != nil && !isURL(it) {
			a.toast(tr("The clipboard doesn't hold a link or a file"))
			return
		}
	}
	a.openItems(items, false)
}

// openURL plays a web link (a direct stream, or a page yt-dlp knows).
func (a *App) openURL(u string) {
	u = strings.TrimSpace(u)
	if u == "" {
		return
	}
	if !isURL(u) && !strings.Contains(u, "://") {
		if _, err := os.Stat(u); err != nil {
			u = "https://" + u
		}
	}
	a.openItems([]string{u}, false)
}

func (a *App) revealCurrent() {
	path := a.player.State.Path
	if path == "" || isURL(path) {
		return
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,", path)
	case "darwin":
		cmd = exec.Command("open", "-R", path)
	default:
		cmd = exec.Command("xdg-open", filepath.Dir(path))
	}
	startDetached(cmd)
}

func (a *App) copyCurrentPath() {
	a.g.SetClipboard(a.player.State.Path)
	a.toast(tr("Copied"))
}

// finished reports whether everything opened has played: the last file
// ended and nothing loops.
func (a *App) finished() bool {
	st := &a.player.State
	return !st.Idle && !st.Loading && st.EOF && st.Paused && !a.casting() &&
		!st.LoopFile && !st.LoopPlaylist && st.PlaylistPos == len(st.Playlist)-1
}

// playAgain plays everything opened from the start.
func (a *App) playAgain() {
	if len(a.player.State.Playlist) > 1 {
		a.player.PlaylistPlay(0)
		a.player.SetPause(false)
		return
	}
	a.player.TogglePause()
}
