package app

import (
	"strconv"

	"github.com/xmhrnx/giode/components/iconbutton"
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/player"
)

func emptyNote(s string) elements.Element {
	return elements.Box(styles.Styles{FlexGrow: 1, Width: -1, Height: -1, Justify: properties.JustifyCenter, Align: properties.AlignCenter},
		elements.Text(s, styles.Styles{Color: colFaint, FontSize: 13, TextAlign: properties.TextAlignCenter}))
}

// playlistRow is one playlist entry; rows are looked up by key (index
// and file), so their hover state survives scrolling.
func (p *sidePanel) playlistRow(i int) elements.Element {
	a := p.u.a
	st := &a.player.State
	if i >= len(st.Playlist) {
		return elements.Box(styles.Styles{})
	}
	e := st.Playlist[i]
	current := i == st.PlaylistPos
	remove := elements.Keyed("playlist-remove", e.Filename+"#"+strconv.Itoa(i), func() *iconbutton.Button {
		return ghostButton("close", 16, tr("Remove"), nil)
	})
	remove.OnClick(func() { elements.AfterLayout(func() { a.player.PlaylistRemove(i) }) })
	return elements.Keyed("playlist-row", e.Filename+"#"+strconv.Itoa(i), func() *elements.Pressable {
		return elements.NewPressable(nil)
	}).Content(func(s elements.PressState) elements.Element {
		bg := ""
		switch {
		case current:
			bg = colSurface3
		case s.Hovered:
			bg = colSurface2
		}
		var num elements.Element = elements.Text(strconv.Itoa(i+1), styles.Styles{Color: colFaint, FontSize: 12})
		if current {
			num = ic("play", 14, colHighlight)
		}
		titleColor := colText
		if !current && !s.Hovered {
			titleColor = "#d4d4da"
		}
		children := []elements.Element{
			elements.Box(styles.Styles{Width: 24, Justify: properties.JustifyCenter}, num),
			elements.Box(styles.Styles{FlexGrow: 1, Width: -1},
				elements.Text(e.Name(), styles.Styles{Color: titleColor, FontSize: 13, MaxLines: 2, TextOverflow: properties.TextOverflowEllipsis})),
		}
		if s.Hovered {
			children = append(children, remove)
		}
		return elements.Row(styles.Styles{Width: -1, Height: -1, Align: properties.AlignCenter, Gap: 8, Wrap: properties.WrapNoWrap,
			Padding: properties.SymmetricInset(8, 4), Background: bg, BorderRadius: 8}, children...)
	}).Label(e.Name()).OnClick(func(elements.ClickEvent) { a.player.PlaylistPlay(i) })
}

func (p *sidePanel) chapterRow(i int) elements.Element {
	a := p.u.a
	st := &a.player.State
	if i >= len(st.Chapters) {
		return elements.Box(styles.Styles{})
	}
	c := st.Chapters[i]
	current := i == st.Chapter
	title := c.Title
	if title == "" {
		title = tr("Chapter {number}", "number", i+1)
	}
	return elements.Keyed("chapter-row", strconv.Itoa(i)+title, func() *elements.Pressable {
		return elements.NewPressable(nil)
	}).Content(func(s elements.PressState) elements.Element {
		bg := ""
		switch {
		case current:
			bg = colSurface3
		case s.Hovered:
			bg = colSurface2
		}
		return elements.Row(styles.Styles{Width: -1, Height: -1, Align: properties.AlignCenter, Gap: 12, Wrap: properties.WrapNoWrap,
			Padding: properties.SymmetricInset(10, 4), Background: bg, BorderRadius: 8},
			elements.Box(styles.Styles{Width: 56}, elements.Text(player.FormatTime(c.Time), styles.Styles{Color: colMuted, FontSize: 12})),
			elements.Box(styles.Styles{FlexGrow: 1, Width: -1}, elements.Text(title, styles.Styles{Color: colText, FontSize: 13, MaxLines: 1, TextOverflow: properties.TextOverflowEllipsis})),
		)
	}).Label(title).OnClick(func(elements.ClickEvent) { a.player.SetChapter(i) })
}
