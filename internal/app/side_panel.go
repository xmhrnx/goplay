package app

import (
	"github.com/xmhrnx/giode/components/iconbutton"
	"github.com/xmhrnx/giode/components/segmented"
	"github.com/xmhrnx/giode/components/virtuallist"
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"
)

const (
	panelPlaylist = iota
	panelChapters
)

// sidePanel slides in from the right with the playlist and chapters.
type sidePanel struct {
	u        *ui
	tabs     *segmented.Control
	playlist *virtuallist.List
	chapters *virtuallist.List
	close    *iconbutton.Button
	add      *iconbutton.Button
	shuffle  *iconbutton.Button
	clear    *iconbutton.Button
}

func newSidePanel(u *ui) *sidePanel {
	a := u.a
	p := &sidePanel{u: u}
	p.tabs = segmented.New(tr("Playlist"), tr("Chapters")).OnChange(func(i int) { u.panelTab = i })
	p.tabs.Styles(styles.Styles{Width: -1, Background: colSurface2, BorderColor: colSurface3, Color: colText, FontSize: 13})
	st := &a.player.State
	p.playlist = virtuallist.New(func() int { return len(st.Playlist) }, 52, gapped(p.playlistRow),
		listStyles()).
		OnActivate(func(i int) { a.player.PlaylistPlay(i) }).
		ClickFocusesRow(false)
	p.chapters = virtuallist.New(func() int { return len(st.Chapters) }, 44, gapped(p.chapterRow),
		listStyles()).
		OnActivate(func(i int) { a.player.SetChapter(i) }).
		ClickFocusesRow(false)
	p.close = ghostButton("close", 18, tr("Close"), func() { u.panelOpen = false })
	p.add = ghostButton("add", 20, tr("Add files"), a.addFilesDialog)
	p.shuffle = ghostButton("shuffle", 20, tr("Shuffle"), func() { a.actions.Run("shuffle") })
	p.clear = ghostButton("delete", 20, tr("Clear the rest"), func() { a.player.PlaylistClear() })
	return p
}

func (p *sidePanel) opened() {
	p.tabs.SetSelected(p.u.panelTab)
	p.scrollToCurrent()
}

func (p *sidePanel) scrollToCurrent() {
	st := &p.u.a.player.State
	if st.PlaylistPos >= 0 {
		p.playlist.ScrollToIndex(st.PlaylistPos, virtuallist.AlignNearest)
	}
	if st.Chapter >= 0 {
		p.chapters.ScrollToIndex(st.Chapter, virtuallist.AlignNearest)
	}
}

func (p *sidePanel) content() elements.Element {
	u := p.u
	st := &u.a.player.State
	p.tabs.SetSelected(u.panelTab)

	header := elements.Row(styles.Styles{Width: -1, Align: properties.AlignCenter, Gap: 8, Wrap: properties.WrapNoWrap},
		elements.Box(styles.Styles{FlexGrow: 1, Width: -1}, p.tabs), p.close)

	var body elements.Element
	var footer elements.Element
	switch u.panelTab {
	case panelChapters:
		if len(st.Chapters) == 0 {
			body = emptyNote(tr("This file has no chapters."))
		} else {
			body = elements.Box(styles.Styles{FlexGrow: 1, Width: -1, Height: -1}, p.chapters)
		}
	default:
		if len(st.Playlist) == 0 {
			body = emptyNote(tr("Nothing queued. Open a folder, or add files."))
		} else {
			body = elements.Box(styles.Styles{FlexGrow: 1, Width: -1, Height: -1}, p.playlist)
		}
		count := tr("{count} items", "count", len(st.Playlist))
		footer = elements.Row(styles.Styles{Width: -1, Align: properties.AlignCenter, Gap: 4, Wrap: properties.WrapNoWrap},
			elements.Box(styles.Styles{FlexGrow: 1, Width: -1}, elements.Text(count, styles.Styles{Color: colMuted, FontSize: 12})),
			p.add, p.shuffle, p.clear)
	}
	children := []elements.Element{header, body}
	if footer != nil {
		children = append(children, footer)
	}
	return elements.Box(styles.Styles{Width: -1, Height: -1, FlexDirection: properties.FlexDirectionColumn, Gap: 12,
		Padding: properties.UniformInset(14)}, children...)
}
