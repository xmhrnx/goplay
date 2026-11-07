package app

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/goplay/internal/library"
)

func isURL(s string) bool { return library.IsURL(s) }

// mediaTypes and subtitleTypes are what the file pickers offer.
func mediaTypes() []string    { return extList(library.MediaExts) }
func subtitleTypes() []string { return extList(library.SubtitleExts) }

func extList(exts map[string]bool) []string {
	var out []string
	for ext := range exts {
		out = append(out, ext)
	}
	return out
}

// pick shows the file picker off the UI goroutine (it blocks) and hands
// the paths back with Post.
func (a *App) pick(types []string, multiple bool, then func([]string)) {
	a.picked(func() ([]string, error) {
		files, err := a.g.PickFiles(context.Background(), giode.PickOptions{Types: types, Multiple: multiple})
		var paths []string
		for _, f := range files {
			f.Close()
			if f.Path != "" {
				paths = append(paths, f.Path)
			}
		}
		return paths, err
	}, then)
}

// picked runs a picker and hands its paths to then on the UI goroutine.
func (a *App) picked(show func() ([]string, error), then func([]string)) {
	go func() {
		paths, err := show()
		a.Post(func() {
			switch {
			case errors.Is(err, giode.ErrCanceled):
			case err != nil:
				a.toastError(tr("The file dialog failed: {error}", "error", err))
			case len(paths) > 0:
				then(paths)
			}
		})
	}()
}

func (a *App) openFileDialog() {
	a.pick(mediaTypes(), true, func(p []string) { a.openItems(p, false) })
}

func (a *App) openFolderDialog() {
	a.picked(func() ([]string, error) {
		d, err := a.g.PickFolder(context.Background(), giode.FolderOptions{Title: tr("Open Folder")})
		return []string{d}, err
	}, func(p []string) { a.openItems(p, false) })
}

// openDisc plays the DVDs and Blu-rays in the drives, or asks for one's
// folder (a copied disc, or its VIDEO_TS or BDMV folder) when there's
// none.
func (a *App) openDisc() {
	go func() {
		discs := library.Discs()
		a.Post(func() {
			if len(discs) > 0 {
				a.openItems(discs, false)
				return
			}
			a.toast(tr("There's no DVD or Blu-ray in a drive. Pick a disc's folder."))
			a.picked(func() ([]string, error) {
				d, err := a.g.PickFolder(context.Background(), giode.FolderOptions{Title: tr("Open DVD or Blu-ray")})
				return []string{d}, err
			}, func(p []string) {
				if k, _ := library.Disc(p[0]); k == library.NoDisc {
					a.toastError(tr("{name} isn't a DVD or Blu-ray", "name", filepath.Base(p[0])))
					return
				}
				a.openItems(p, false)
			})
		})
	}()
}

func (a *App) addFilesDialog() {
	a.pick(mediaTypes(), true, func(p []string) { a.openItems(p, true) })
}

func (a *App) loadSubtitleDialog() {
	a.pick(subtitleTypes(), false, func(p []string) { a.addSubtitle(p[0]) })
}

func (a *App) loadAudioDialog() {
	a.pick(mediaTypes(), false, func(p []string) {
		if err := a.player.AddAudio(p[0]); err != nil {
			a.toastError(tr("Couldn't load the audio track: {error}", "error", err))
		}
	})
}

func (a *App) addSubtitle(path string) {
	if err := a.player.AddSubtitle(path); err != nil {
		a.toastError(tr("Couldn't load the subtitles: {error}", "error", err))
		return
	}
	a.osd(tr("Subtitles: {track}", "track", filepath.Base(path)))
}
