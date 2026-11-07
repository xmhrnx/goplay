package app

import (
	"context"
	"strings"
	"time"

	"github.com/xmhrnx/giode/components/virtuallist"

	"github.com/mehran9675/goplay/internal/cast"
)

func sortServers(ss []*cast.MediaServer) {
	for i := 1; i < len(ss); i++ {
		for j := i; j > 0 && strings.ToLower(ss[j].Name) < strings.ToLower(ss[j-1].Name); j-- {
			ss[j], ss[j-1] = ss[j-1], ss[j]
		}
	}
}

// openServer starts browsing s at its top folder.
func (n *netMedia) openServer(s *cast.MediaServer) {
	n.stopScan()
	n.server = s
	n.path = []netFolder{{id: cast.RootID, title: s.Name}}
	n.load()
}

// load lists the folder at the end of the path.
func (n *netMedia) load() {
	a := n.u.a
	n.gen++
	gen, s, id := n.gen, n.server, n.path[len(n.path)-1].id
	n.loading, n.err, n.entries = true, "", nil
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		entries, err := s.Browse(ctx, id)
		a.Post(func() {
			if gen != n.gen {
				return
			}
			n.loading = false
			if err != nil {
				n.err = tr("Couldn't open this folder: {error}", "error", err)
				return
			}
			n.entries = entries
			n.list.ScrollToIndex(0, virtuallist.AlignStart)
		})
	}()
}

// up goes to the parent folder, or back to the list of servers.
func (n *netMedia) up() {
	if len(n.path) > 1 {
		n.path = n.path[:len(n.path)-1]
		n.load()
		return
	}
	n.gen++
	n.server, n.path, n.entries, n.err, n.loading = nil, nil, nil, "", false
	if len(n.servers) == 0 {
		n.scan()
	}
}

// activate opens a folder, or plays the folder's files from entry i.
func (n *netMedia) activate(i int) {
	if i < 0 || i >= len(n.entries) {
		return
	}
	if e := n.entries[i]; e.Folder {
		n.path = append(n.path, netFolder{id: e.ID, title: e.Title})
		n.load()
		return
	}
	n.play(i)
}

// play plays the listed files from entry i on, as a playlist.
func (n *netMedia) play(from int) {
	var urls []string
	for _, e := range n.entries[from:] {
		if !e.Folder {
			urls = append(urls, e.URL)
			n.u.a.player.Name(e.URL, e.Title)
		}
	}
	if len(urls) == 0 {
		return
	}
	n.open = false
	n.u.a.openItems(urls, false)
}

func (n *netMedia) files() int {
	c := 0
	for _, e := range n.entries {
		if !e.Folder {
			c++
		}
	}
	return c
}
