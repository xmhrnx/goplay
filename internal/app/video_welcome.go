package app

import (
	"path/filepath"
	"strings"

	"github.com/mehran9675/giode/components/button"
	"github.com/mehran9675/giode/components/progress"
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/library"
	"github.com/mehran9675/goplay/internal/player"
)

// welcomeScreen is the empty state: what to open, and what was played
// lately. Once everything has played it's the finished screen, over the
// last frame.
type welcomeScreen struct {
	u                                  *ui
	openBtn, urlBtn, folderBtn, netBtn *button.Button
	// againBtn and openNextBtn are the finished screen's: there the
	// replay is the main action and opening another file comes second.
	againBtn, openNextBtn *button.Button
}

func newWelcomeScreen(u *ui) *welcomeScreen {
	a := u.a
	w := &welcomeScreen{u: u}
	primary := styles.Styles{
		Background: colHighlight, Color: colOnLight, FontSize: 14, FontWeight: properties.FontWeightSemiBold,
		Padding: properties.SymmetricInset(18, 11), BorderRadius: 22,
	}
	secondary := styles.Styles{
		Background: colSurface3, Color: colText, FontSize: 14, FontWeight: properties.FontWeightMedium,
		Padding: properties.SymmetricInset(18, 11), BorderRadius: 22,
	}
	w.openBtn = button.New(tr("Open File"), primary).Icon(ic("file", 18, colOnLight)).OnClick(a.openFileDialog)
	w.urlBtn = button.New(tr("Open URL"), secondary).Icon(ic("link", 18, colText)).OnClick(func() { u.dialogs.url.show() })
	w.folderBtn = button.New(tr("Open Folder"), secondary).Icon(ic("folder", 18, colText)).OnClick(a.openFolderDialog)
	w.netBtn = button.New(tr("Network"), secondary).Icon(ic("tv", 18, colText)).OnClick(func() { u.dialogs.net.show() })
	w.againBtn = button.New(tr("Play Again"), primary).Icon(ic("replay", 18, colOnLight)).OnClick(a.playAgain)
	w.openNextBtn = button.New(tr("Open File"), secondary).Icon(ic("file", 18, colText)).OnClick(a.openFileDialog)
	return w
}

// element is the screen; finished says so and offers to play it again.
func (w *welcomeScreen) element(finished bool) elements.Element {
	a := w.u.a
	head := elements.Box(styles.Styles{FlexDirection: properties.FlexDirectionColumn, Align: properties.AlignCenter, Gap: 10},
		ic("logo", 56, colHighlight),
		elements.Text(tr("Drop a video here, or"), styles.Styles{Color: colMuted, FontSize: 15}),
		elements.Row(styles.Styles{Gap: 10, Wrap: properties.WrapNoWrap, Margin: properties.Inset{Top: 6}}, w.openBtn, w.urlBtn, w.folderBtn, w.netBtn),
	)
	if finished {
		st := &a.player.State
		note := displayTitle(st)
		if n := len(st.Playlist); n > 1 {
			note = tr("All {count} files in the playlist have played.", "count", n)
		}
		head = elements.Box(styles.Styles{FlexDirection: properties.FlexDirectionColumn, Align: properties.AlignCenter, Gap: 10},
			elements.Text(tr("Finished playing"), styles.Styles{Color: colText, FontSize: 20, FontWeight: properties.FontWeightSemiBold}),
			elements.Box(styles.Styles{MaxWidth: 520},
				elements.Text(note, styles.Styles{Color: colMuted, FontSize: 14, MaxLines: 2, TextAlign: properties.TextAlignCenter})),
			elements.Box(styles.Styles{Margin: properties.Inset{Top: 6, Bottom: 10}}, w.againBtn),
			elements.Text(tr("Or open something else"), styles.Styles{Color: colMuted, FontSize: 13}),
			elements.Row(styles.Styles{Gap: 10, Wrap: properties.WrapNoWrap}, w.openNextBtn, w.urlBtn, w.folderBtn, w.netBtn),
		)
	}
	col := []elements.Element{head}
	if recent := a.history.Items(); len(recent) > 0 {
		// The finished screen's heading is taller; it lists fewer.
		limit := 6
		if finished {
			limit = 3
		}
		if len(recent) > limit {
			recent = recent[:limit]
		}
		rows := []elements.Element{elements.Row(styles.Styles{Width: -1, Align: properties.AlignCenter, Wrap: properties.WrapNoWrap},
			elements.Box(styles.Styles{FlexGrow: 1, Width: -1}, elements.Text(tr("Recent"), styles.Styles{Color: colMuted, FontSize: 12, FontWeight: properties.FontWeightSemiBold})),
			w.clearRecent(),
		)}
		for _, r := range recent {
			rows = append(rows, w.recentRow(r))
		}
		col = append(col, elements.Box(styles.Styles{Width: 460, FlexDirection: properties.FlexDirectionColumn, Gap: 2, Margin: properties.Inset{Top: 28}}, rows...))
	}
	screen := centered(elements.Box(styles.Styles{FlexDirection: properties.FlexDirectionColumn, Align: properties.AlignCenter}, col...))
	if finished {
		// The last frame stays behind, dimmed.
		return elements.Stack(styles.Styles{Width: -1, Height: -1},
			elements.Box(styles.Styles{Width: -1, Height: -1, Background: "#000000d9"}),
			screen)
	}
	return screen
}

func (w *welcomeScreen) clearRecent() elements.Element {
	a := w.u.a
	return elements.Keyed("welcome", "clear", func() *elements.Pressable {
		return elements.NewPressable(func(s elements.PressState) elements.Element {
			c := colFaint
			if s.Hovered {
				c = colText
			}
			return elements.Text(tr("Clear"), styles.Styles{Color: c, FontSize: 12})
		}).Label(tr("Clear recent")).Focusable(false).OnClick(func(elements.ClickEvent) {
			a.history.Clear()
			a.history.Save()
		})
	})
}

// recentRow is one recently played file: name, where it's from, and how
// far it got.
func (w *welcomeScreen) recentRow(r library.Recent) elements.Element {
	a := w.u.a
	return elements.Keyed("recent", r.Path, func() *elements.Pressable {
		return elements.NewPressable(func(s elements.PressState) elements.Element {
			bg := ""
			if s.Hovered {
				bg = colSurface2
			}
			icon := "file"
			where := filepath.Dir(r.Path)
			if library.IsURL(r.Path) {
				icon, where = "link", hostOf(r.Path)
			}
			meta := where
			if p := r.Progress(); p > 0.02 && p < 0.98 {
				meta = tr("{time} left · {where}", "time", player.FormatTime(r.Duration-r.Position), "where", where)
			}
			body := []elements.Element{
				elements.Text(r.Name(), styles.Styles{Color: colText, FontSize: 13, MaxLines: 1, TextOverflow: properties.TextOverflowEllipsis}),
				elements.Text(meta, styles.Styles{Color: colFaint, FontSize: 11, MaxLines: 1, TextOverflow: properties.TextOverflowEllipsis}),
			}
			if p := r.Progress(); p > 0.02 && p < 0.98 {
				body = append(body, progress.Progress(float32(p), styles.Styles{Width: -1, Height: 2, Color: colMuted, Background: colSurface3}))
			}
			return elements.Row(styles.Styles{Width: -1, Align: properties.AlignCenter, Gap: 12, Padding: properties.SymmetricInset(10, 8),
				Background: bg, BorderRadius: 8, Wrap: properties.WrapNoWrap},
				ic(icon, 18, colMuted),
				elements.Box(styles.Styles{FlexGrow: 1, Width: -1, FlexDirection: properties.FlexDirectionColumn, Gap: 3}, body...),
			)
		}).Label(r.Name()).OnClick(func(elements.ClickEvent) { a.openItems([]string{r.Path}, false) })
	})
}

// hostOf is a link's site, for the recent list.
func hostOf(u string) string {
	s := u
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimPrefix(s, "www.")
}
