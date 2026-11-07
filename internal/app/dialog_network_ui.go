package app

import (
	"strings"

	"github.com/xmhrnx/giode/components/spinner"
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/cast"
	"github.com/mehran9675/goplay/internal/player"
)

func (n *netMedia) content() elements.Element {
	if n.server == nil {
		return n.serversContent()
	}
	crumbs := make([]string, len(n.path))
	for i, f := range n.path {
		crumbs[i] = f.title
	}
	rows := []elements.Element{heading(n.path[len(n.path)-1].title)}
	if len(crumbs) > 1 {
		rows = append(rows, note(strings.Join(crumbs, "  ›  ")))
	}
	switch {
	case n.loading:
		rows = append(rows, busy(tr("Opening…")))
	case n.err != "":
		rows = append(rows, note(n.err))
	case len(n.entries) == 0:
		rows = append(rows, note(tr("There's no video or music in this folder.")))
	default:
		h := min(len(n.entries), 8) * 52
		rows = append(rows, elements.Box(styles.Styles{Width: -1, Height: properties.Height(h)}, n.list))
	}
	btns := []elements.Element{n.back}
	btns = append(btns, elements.Box(styles.Styles{FlexGrow: 1, Width: -1}))
	if !n.loading && n.files() > 0 {
		btns = append(btns, n.playAll)
	}
	btns = append(btns, n.closeBtn)
	rows = append(rows, elements.Row(styles.Styles{Width: -1, Gap: 8, Align: properties.AlignCenter, Wrap: properties.WrapNoWrap,
		Margin: properties.Inset{Top: 6}}, btns...))
	return column(10, rows...)
}

func (n *netMedia) serversContent() elements.Element {
	rows := []elements.Element{
		heading(tr("Media on your network")),
		note(tr("Videos and music shared by computers, NAS drives and media servers (DLNA/UPnP) on this network.")),
	}
	for _, s := range n.servers {
		rows = append(rows, n.serverRow(s))
	}
	if n.scanning {
		rows = append(rows, busy(tr("Looking for media servers…")))
	} else if len(n.servers) == 0 {
		rows = append(rows, note(tr("No media servers found. Make sure sharing is on and they're on the same network as this computer, "+
			"or add a server's address below.")))
	}
	for _, addr := range n.silent {
		rows = append(rows, note(tr("No media server answered at {address}. Saved addresses are in Settings › Casting.", "address", addr)))
	}
	rows = append(rows,
		elements.Row(styles.Styles{Width: -1, Gap: 8, Align: properties.AlignCenter, Wrap: properties.WrapNoWrap, Margin: properties.Inset{Top: 6}},
			elements.Box(styles.Styles{FlexGrow: 1, Width: -1}, n.addrInput), n.addBtn),
		buttonsRow(n.rescan, n.closeBtn))
	return column(10, rows...)
}

func busy(text string) elements.Element {
	return elements.Row(styles.Styles{Gap: 10, Align: properties.AlignCenter, Padding: properties.SymmetricInset(4, 6)},
		spinner.Spinner(styles.Styles{Width: 18, Height: 18, Color: colMuted}), note(text))
}

func (n *netMedia) serverRow(s *cast.MediaServer) elements.Element {
	return elements.Keyed("net-server", s.ID, func() *elements.Pressable {
		return elements.NewPressable(func(ps elements.PressState) elements.Element {
			bg := colSurface2
			if ps.Hovered {
				bg = colSurface3
			}
			children := []elements.Element{elements.Text(s.Name, styles.Styles{Color: colText, FontSize: 14, FontWeight: properties.FontWeightMedium})}
			if s.Model != "" {
				children = append(children, elements.Text(s.Model, styles.Styles{Color: colFaint, FontSize: 12}))
			}
			return elements.Row(styles.Styles{Width: -1, Gap: 12, Align: properties.AlignCenter, Padding: properties.SymmetricInset(12, 10),
				Background: bg, BorderRadius: 10, Wrap: properties.WrapNoWrap},
				ic("folder", 22, colText),
				elements.Box(styles.Styles{FlexGrow: 1, Width: -1, FlexDirection: properties.FlexDirectionColumn, Gap: 2}, children...))
		}).Label(s.Name).OnClick(func(elements.ClickEvent) { n.openServer(s) })
	})
}

// row is entry i of the folder: a folder to open or a file to play.
func (n *netMedia) row(i int) elements.Element {
	if i >= len(n.entries) {
		return elements.Box(styles.Styles{})
	}
	e := n.entries[i]
	return elements.Keyed("net-entry", n.server.ID+"\x00"+e.ID, func() *elements.Pressable {
		return elements.NewPressable(func(ps elements.PressState) elements.Element {
			var bg properties.Color
			if ps.Hovered {
				bg = colSurface2
			}
			icon, sub := "file", ""
			switch {
			case e.Folder:
				icon = "folder"
				if e.Count >= 0 {
					sub = tr("{count} items", "count", e.Count)
				}
			case e.Audio && !e.Video:
				icon = "audio"
				sub = strings.Join(nonEmpty(e.Artist, e.Album), " · ")
			}
			if !e.Folder && e.Duration > 0 {
				sub = strings.Join(nonEmpty(sub, player.FormatTime(e.Duration)), " · ")
			}
			text := []elements.Element{elements.Text(e.Title, styles.Styles{Color: colText, FontSize: 13, MaxLines: 1, TextOverflow: properties.TextOverflowEllipsis})}
			if sub != "" {
				text = append(text, elements.Text(sub, styles.Styles{Color: colFaint, FontSize: 11, MaxLines: 1}))
			}
			row := []elements.Element{ic(icon, 20, colMuted),
				elements.Box(styles.Styles{FlexGrow: 1, Width: -1, FlexDirection: properties.FlexDirectionColumn, Gap: 2}, text...)}
			if e.Folder {
				row = append(row, ic("chevron", 18, colFaint))
			}
			return elements.Row(styles.Styles{Width: -1, Height: 52, Gap: 12, Align: properties.AlignCenter, Padding: properties.SymmetricInset(10, 0),
				Background: bg, BorderRadius: 8, Wrap: properties.WrapNoWrap}, row...)
		}).Label(e.Title).OnClick(func(elements.ClickEvent) {
			// The list may have changed since this row was built.
			if idx := n.indexOf(e.ID); idx >= 0 {
				n.activate(idx)
			}
		})
	})
}

func (n *netMedia) indexOf(id string) int {
	for i := range n.entries {
		if n.entries[i].ID == id {
			return i
		}
	}
	return -1
}

func nonEmpty(ss ...string) []string {
	var out []string
	for _, s := range ss {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
