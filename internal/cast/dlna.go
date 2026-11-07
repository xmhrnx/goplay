package cast

import (
	"bufio"
	"bytes"
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

// DLNA/UPnP: media renderers answer SSDP searches with the URL of an XML
// description listing their services; playback is driven with SOAP
// calls to the AVTransport service, volume with RenderingControl.

// DLNARenderer is a UPnP media renderer.
type DLNARenderer struct {
	id, name, model, host       string
	transportURL, transportType string
	renderingURL, renderingType string
	mu                          sync.Mutex
	duration                    float64
}

const (
	ssdpAddr        = "239.255.255.250:1900"
	rendererType    = "urn:schemas-upnp-org:device:MediaRenderer:1"
	avTransport     = "urn:schemas-upnp-org:service:AVTransport:1"
	renderingContrl = "urn:schemas-upnp-org:service:RenderingControl:1"
)

func discoverDLNA(ctx context.Context, report func(Device)) {
	ssdpSearch(ctx, ssdpAddr, rendererType, func(loc string) {
		if d, err := describeRenderer(ctx, loc); err == nil {
			report(d)
		}
	})
}

// ssdpSearch asks for UPnP devices of type st every few seconds until
// ctx ends, calling found on its own goroutine with each new device's
// description URL. dst is the multicast group (ssdpAddr), or one
// device's address and SSDP port to ask it alone.
func ssdpSearch(ctx context.Context, dst, st string, found func(location string)) {
	to, err := net.ResolveUDPAddr("udp4", dst)
	if err != nil {
		return
	}
	search := []byte("M-SEARCH * HTTP/1.1\r\nHOST: " + dst + "\r\nMAN: \"ssdp:discover\"\r\nMX: 2\r\nST: " + st + "\r\n\r\n")
	seen := map[string]bool{}
	query(ctx, to, search, 4*time.Second, func(p packet) {
		resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(p.data)), nil)
		if err != nil {
			return
		}
		loc := resp.Header.Get("Location")
		resp.Body.Close()
		if loc == "" || seen[loc] {
			return
		}
		seen[loc] = true
		go found(loc)
	})
}

func (d *DLNARenderer) ID() string    { return d.id }
func (d *DLNARenderer) Name() string  { return d.name }
func (d *DLNARenderer) Kind() string  { return "DLNA" }
func (d *DLNARenderer) Model() string { return d.model }

// Addr is the device's network address, for the media server.
func (d *DLNARenderer) Addr() string { return d.host }
