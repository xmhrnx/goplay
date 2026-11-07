package app

import (
	"context"
	"slices"
	"strings"
	"sync"
	"time"

	"gioui.org/layout"

	"github.com/xmhrnx/giode/components/button"
	"github.com/xmhrnx/giode/components/dialog"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/virtuallist"

	"github.com/mehran9675/goplay/internal/cast"
	"github.com/mehran9675/goplay/internal/settings"
)

// netMedia browses the media servers on the network (DLNA/UPnP: a NAS,
// Plex, Jellyfin, Windows' media sharing) and plays what they share.
type netMedia struct {
	modal
	u    *ui
	list *virtuallist.List

	scanMu     sync.Mutex
	scanCancel context.CancelFunc
	scanning   bool
	// scanGen tells a scan's results from those of one stopped before.
	scanGen int
	servers []*cast.MediaServer
	// silent are the saved addresses where no server answered.
	silent []string

	// server is the one being browsed, path the folders from its top
	// down to the one listed, entries that folder's contents.
	server  *cast.MediaServer
	path    []netFolder
	entries []cast.MediaEntry
	loading bool
	err     string
	// gen tells a listing that arrives late from one still wanted.
	gen int

	addrInput                               *input.Input
	back, rescan, playAll, closeBtn, addBtn *button.Button
}

type netFolder struct{ id, title string }

func newNetMedia(u *ui) *netMedia {
	n := &netMedia{u: u}
	n.list = virtuallist.New(func() int { return len(n.entries) }, 52, gapped(n.row), listStyles()).
		OnActivate(n.activate).
		ClickFocusesRow(false)
	n.back = secondaryButton(tr("Back"), n.up)
	n.rescan = secondaryButton(tr("Search Again"), n.scan)
	n.playAll = primaryButton(tr("Play All"), func() { n.play(0) })
	n.closeBtn = secondaryButton(tr("Close"), func() { n.open = false })
	n.addrInput = input.New(tr("A server's address, like 192.168.1.20 or 192.168.1.20:8096"), fieldStyles()).Keyboard(input.KeyboardURL).
		Submit(n.addAddress)
	n.addBtn = secondaryButton(tr("Add"), func() { n.addAddress(n.addrInput.Text()) })
	n.dlg = dialog.New(&n.open, n.content, dialogStyles(600))
	return n
}

// show opens the window, looking for servers when none is being
// browsed.
func (n *netMedia) show() {
	n.open = true
	if n.server == nil {
		n.scan()
	}
}

func (n *netMedia) layout(gtx layout.Context) {
	n.dlg.Layout(gtx)
	if !n.open && n.scanning {
		n.stopScan()
	}
}

// scan looks for media servers for a while; they arrive through Post.
func (n *netMedia) scan() {
	a := n.u.a
	n.scanMu.Lock()
	if n.scanCancel != nil {
		n.scanCancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	n.scanCancel = cancel
	n.scanMu.Unlock()
	n.scanning = true
	n.scanGen++
	gen := n.scanGen
	n.servers, n.silent = nil, nil
	n.server, n.path, n.entries, n.err = nil, nil, nil, ""
	addrs := n.addresses()
	go func() {
		silent := cast.DiscoverMediaServers(ctx, addrs, func(s *cast.MediaServer) {
			a.Post(func() {
				if gen != n.scanGen || slices.ContainsFunc(n.servers, s.Same) {
					return
				}
				n.servers = append(n.servers, s)
				sortServers(n.servers)
			})
		})
		a.Post(func() {
			if gen != n.scanGen {
				return
			}
			// A scan stopped early didn't give the addresses their time.
			if ctx.Err() == context.DeadlineExceeded {
				n.silent = silent
			}
			n.scanning = false
		})
	}()
}

// addresses are the media server addresses saved in Settings.
func (n *netMedia) addresses() []string {
	var out []string
	for _, line := range strings.Split(n.u.a.settings.Get(settings.KeyMediaServers), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return out
}

// addAddress saves a media server's address and searches again.
func (n *netMedia) addAddress(addr string) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return
	}
	addrs := n.addresses()
	for _, a := range addrs {
		if strings.EqualFold(a, addr) {
			n.addrInput.SetText("")
			n.scan()
			return
		}
	}
	st := n.u.a.settings
	st.Set(settings.KeyMediaServers, strings.Join(append(addrs, addr), "\n"))
	st.Flush()
	n.addrInput.SetText("")
	n.scan()
}

func (n *netMedia) stopScan() {
	n.scanMu.Lock()
	if n.scanCancel != nil {
		n.scanCancel()
		n.scanCancel = nil
	}
	n.scanMu.Unlock()
	n.scanning = false
}
