package app

import (
	"path/filepath"
	"strings"

	"github.com/xmhrnx/giode/components/progress"
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/library"
	"github.com/mehran9675/goplay/internal/player"
)

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
