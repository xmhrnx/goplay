package cast

import (
	"bufio"
	"context"
	"net"
	"strconv"
	"sync"
)

// Google Cast: devices are found by mDNS (_googlecast._tcp) and spoken to
// over TLS on port 8009 with length-prefixed protobuf CastMessages whose
// payloads are JSON. goplay launches the Default Media Receiver, which
// plays a URL it fetches itself.

// Chromecast is a Google Cast device.
type Chromecast struct {
	id, name, model, addr string

	mu        sync.Mutex
	conn      net.Conn
	w         *bufio.Writer
	requestID int
	waiters   map[int]chan map[string]any
	transport string // the media app's transport ID
	session   string // the media app's session ID
	mediaID   int    // media session ID
	status    Status
	closed    chan struct{}
}

const (
	nsConnection = "urn:x-cast:com.google.cast.tp.connection"
	nsHeartbeat  = "urn:x-cast:com.google.cast.tp.heartbeat"
	nsReceiver   = "urn:x-cast:com.google.cast.receiver"
	nsMedia      = "urn:x-cast:com.google.cast.media"

	defaultReceiverApp = "CC1AD845"
	senderID           = "sender-goplay"
	receiverID         = "receiver-0"
)

func discoverChromecasts(ctx context.Context, report func(Device)) {
	browse(ctx, "_googlecast._tcp.local.", func(s service) {
		name := s.TXT["fn"]
		if name == "" {
			name = s.Instance
		}
		id := s.TXT["id"]
		if id == "" {
			id = s.Instance
		}
		report(&Chromecast{
			id:    "cast:" + id,
			name:  name,
			model: s.TXT["md"],
			addr:  net.JoinHostPort(s.Addr.String(), strconv.Itoa(s.Port)),
		})
	})
}

func (c *Chromecast) ID() string    { return c.id }
func (c *Chromecast) Name() string  { return c.name }
func (c *Chromecast) Kind() string  { return "Chromecast" }
func (c *Chromecast) Model() string { return c.model }

// Addr is the device's network address, for the media server.
func (c *Chromecast) Addr() string { return c.addr }
