package cast

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// DLNA media servers (a NAS, Plex, Jellyfin, Windows Media Player's
// sharing, a router with a USB disk) answer SSDP searches like renderers
// do, and list their library through the ContentDirectory service: a
// tree of containers (folders) and items, each item with HTTP URLs mpv
// plays directly.

// MediaServer is a DLNA/UPnP media server on the network.
type MediaServer struct {
	ID, Name, Model string
	browseURL       string
	browseType      string
}

// MediaEntry is a folder or a playable file on a media server.
type MediaEntry struct {
	ID, Title string
	// Folder is a container to browse into; Count is how many entries
	// it holds, or -1 when the server doesn't say.
	Folder bool
	Count  int
	// URL plays the file; Video and Audio say what it is.
	URL          string
	Video, Audio bool
	// Duration is in seconds, 0 when unknown.
	Duration float64
	// Artist and Album are set for music.
	Artist, Album string
}

const (
	serverType       = "urn:schemas-upnp-org:device:MediaServer:1"
	contentDirPrefix = "urn:schemas-upnp-org:service:ContentDirectory:"
	// browsePage is how many entries one Browse call asks for; servers
	// may return fewer.
	browsePage = 200
	// maxEntries bounds a folder's listing.
	maxEntries = 5000
)

// embyPort is where Emby and Jellyfin serve their web pages and API by
// default.
const embyPort = "8096"

// Same reports whether s and o are the same server.
func (s *MediaServer) Same(o *MediaServer) bool {
	return s.key() == o.key() || s.browseURL == o.browseURL
}

// key identifies the server: its ID, with case and dashes ignored. Emby
// gives the same UUID with dashes over SSDP and without through its web
// API, each at its own description URL.
func (s *MediaServer) key() string {
	return strings.ToLower(strings.ReplaceAll(s.ID, "-", ""))
}

// DiscoverMediaServers looks for media servers until ctx ends, on the
// whole network and at each of addrs (see askAddress), calling found
// from its own goroutines with each one, once. It returns the addresses
// where none answered.
func DiscoverMediaServers(ctx context.Context, addrs []string, found func(*MediaServer)) (silent []string) {
	var (
		mu       sync.Mutex
		seen     = map[string]bool{}
		answered = make([]bool, len(addrs))
		wg       sync.WaitGroup
	)
	add := func(loc string) bool {
		s, err := describeServer(ctx, loc)
		if err != nil {
			return false
		}
		// A server answering on several interfaces, or both to the
		// search and at a saved address, is listed once.
		mu.Lock()
		dup := seen[s.key()] || seen[s.browseURL]
		seen[s.key()], seen[s.browseURL] = true, true
		mu.Unlock()
		if !dup {
			found(s)
		}
		return true
	}
	for i, addr := range addrs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			askAddress(ctx, addr, func(loc string) {
				if add(loc) {
					mu.Lock()
					answered[i] = true
					mu.Unlock()
				}
			})
		}()
	}
	ssdpSearch(ctx, ssdpAddr, serverType, func(loc string) { add(loc) })
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	for i, addr := range addrs {
		if !answered[i] {
			silent = append(silent, addr)
		}
	}
	return silent
}

// askAddress asks one device for its media servers until ctx ends,
// calling found with each one's description URL. addr is a host name or
// IP address, with or without the port of its web page, or the URL of a
// server's UPnP description. The device is asked by SSDP directly, which
// reaches servers a multicast search doesn't (a firewall or router that
// drops multicast, another subnet); Emby and Jellyfin, whose description
// is at a URL with their server ID in it, are also asked through their
// web API.
func askAddress(ctx context.Context, addr string, found func(location string)) {
	addr = strings.TrimSpace(addr)
	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}
	u, err := url.Parse(addr)
	if err != nil || u.Hostname() == "" {
		return
	}
	if u.Path != "" && u.Path != "/" {
		found(u.String())
		return
	}
	go func() {
		if loc := embyDescription(ctx, u); loc != "" {
			found(loc)
		}
	}()
	ssdpSearch(ctx, net.JoinHostPort(u.Hostname(), "1900"), serverType, found)
}

// embyDescription returns the UPnP description URL of the Emby or
// Jellyfin server at u (embyPort when u has no port), or "" when there's
// none: their public system information gives the server's ID.
func embyDescription(ctx context.Context, u *url.URL) string {
	host := u.Host
	if u.Port() == "" {
		host = net.JoinHostPort(u.Hostname(), embyPort)
	}
	base := u.Scheme + "://" + host
	// Emby answers under /emby too; Jellyfin only at the top.
	for _, path := range []string{"/System/Info/Public", "/emby/System/Info/Public"} {
		var info struct{ Id string }
		if getJSON(ctx, base+path, &info) == nil && info.Id != "" {
			return base + "/dlna/" + url.PathEscape(info.Id) + "/description.xml"
		}
	}
	return ""
}

func getJSON(ctx context.Context, link string, v any) error {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cast: %s: %s", link, resp.Status)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(v)
}
