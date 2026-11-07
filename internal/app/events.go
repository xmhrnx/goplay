package app

import (
	"path/filepath"
	"strings"

	"github.com/mehran9675/goplay/internal/library"
	"github.com/mehran9675/goplay/internal/mpv"
	"github.com/mehran9675/goplay/internal/player"
)

// fileLoaded runs when mpv starts playing a file.
func (a *App) fileLoaded() {
	st := &a.player.State
	a.history.Played(library.Recent{Path: st.Path, Title: titleFor(st)})
	a.history.Save()
	a.applyTrackChoices()
	a.stereoFileLoaded()
	a.ui.fileLoaded()
	if cmds := a.startupCommands; cmds != nil {
		a.startupCommands = nil
		a.Post(func() { a.runCommands(cmds) })
	}
}

// fileEnded records how far the file got, and reports failures.
func (a *App) fileEnded(reason int, err error) {
	st := &a.player.State
	if a.lastPath != "" && st.Duration > 0 {
		pos := st.Position
		if reason == mpv.EndEOF {
			pos = st.Duration
		}
		a.history.Update(a.lastPath, a.lastTitle, pos, st.Duration)
		a.history.Save()
	}
	if cmds := a.startupCommands; cmds != nil && reason == mpv.EndError {
		a.startupCommands = nil
		a.Post(func() { a.runCommands(cmds) })
	}
	if reason == mpv.EndError {
		path := a.player.State.Path
		name := filepath.Base(path)
		if library.IsDiscURL(path) {
			name = library.DiscName(path)
		}
		if name == "." || name == "" {
			name = tr("the file")
		}
		msg := tr("Couldn't play {name}", "name", name)
		if err != nil {
			msg += ": " + err.Error()
		}
		if strings.HasPrefix(path, "bd://") {
			// mpv reads Blu-rays with libbluray, which can't decrypt one
			// by itself.
			msg += ". " + tr("Most Blu-rays sold are encrypted and need libaacs and a key database to play.")
		}
		a.toastError(msg)
	}
}

func titleFor(st *player.State) string {
	if st.Title != "" && st.Title != filepath.Base(st.Path) {
		return st.Title
	}
	return ""
}

// appendLog keeps the last mpv log lines for the log window.
func (a *App) appendLog(level, prefix, text string) {
	line := "[" + prefix + "] " + level + ": " + trimNewline(text)
	a.logMu.Lock()
	a.log = append(a.log, line)
	if len(a.log) > 2000 {
		a.log = append(a.log[:0], a.log[len(a.log)-1500:]...)
	}
	a.logMu.Unlock()
}

func trimNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

func (a *App) logLines() []string {
	a.logMu.Lock()
	defer a.logMu.Unlock()
	return append([]string(nil), a.log...)
}
