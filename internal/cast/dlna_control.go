package cast

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"
)

func (d *DLNARenderer) Load(ctx context.Context, m Media) error {
	d.mu.Lock()
	d.duration = m.Duration
	d.mu.Unlock()
	meta := didl(m)
	if _, err := d.soap(ctx, d.transportURL, d.transportType, "SetAVTransportURI",
		"<InstanceID>0</InstanceID><CurrentURI>"+html.EscapeString(m.URL)+"</CurrentURI><CurrentURIMetaData>"+html.EscapeString(meta)+"</CurrentURIMetaData>"); err != nil {
		return err
	}
	if err := d.Play(); err != nil {
		return err
	}
	if m.Start > 1 {
		// Many renderers only accept a seek once playing.
		go func() {
			time.Sleep(2 * time.Second)
			d.Seek(m.Start)
		}()
	}
	return nil
}

// didl is the DIDL-Lite metadata renderers want with a URI: the title,
// the resource's type, and subtitles in the forms TVs look for.
func didl(m Media) string {
	var b strings.Builder
	b.WriteString(`<DIDL-Lite xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/" xmlns:dc="http://purl.org/dc/elements/1.1/" ` +
		`xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns:sec="http://www.sec.co.kr/">`)
	b.WriteString(`<item id="0" parentID="-1" restricted="1"><dc:title>`)
	b.WriteString(html.EscapeString(m.Title))
	b.WriteString(`</dc:title><upnp:class>object.item.videoItem</upnp:class>`)
	fmt.Fprintf(&b, `<res protocolInfo="http-get:*:%s:DLNA.ORG_OP=01;DLNA.ORG_FLAGS=01700000000000000000000000000000">%s</res>`,
		html.EscapeString(m.ContentType), html.EscapeString(m.URL))
	if m.Subtitles != "" {
		fmt.Fprintf(&b, `<res protocolInfo="http-get:*:text/vtt:*">%s</res>`, html.EscapeString(m.Subtitles))
		fmt.Fprintf(&b, `<sec:CaptionInfoEx sec:type="vtt">%s</sec:CaptionInfoEx>`, html.EscapeString(m.Subtitles))
	}
	b.WriteString(`</item></DIDL-Lite>`)
	return b.String()
}

func (d *DLNARenderer) Play() error {
	_, err := d.soap(context.Background(), d.transportURL, d.transportType, "Play", "<InstanceID>0</InstanceID><Speed>1</Speed>")
	return err
}

func (d *DLNARenderer) Pause() error {
	_, err := d.soap(context.Background(), d.transportURL, d.transportType, "Pause", "<InstanceID>0</InstanceID>")
	return err
}

func (d *DLNARenderer) Seek(s float64) error {
	_, err := d.soap(context.Background(), d.transportURL, d.transportType, "Seek",
		"<InstanceID>0</InstanceID><Unit>REL_TIME</Unit><Target>"+hms(s)+"</Target>")
	return err
}

func (d *DLNARenderer) Stop() error {
	_, err := d.soap(context.Background(), d.transportURL, d.transportType, "Stop", "<InstanceID>0</InstanceID>")
	return err
}

func (d *DLNARenderer) SetVolume(level float64) error {
	if d.renderingURL == "" {
		return ErrUnsupported
	}
	v := int(max(0, min(1, level)) * 100)
	_, err := d.soap(context.Background(), d.renderingURL, d.renderingType, "SetVolume",
		"<InstanceID>0</InstanceID><Channel>Master</Channel><DesiredVolume>"+strconv.Itoa(v)+"</DesiredVolume>")
	return err
}

func (d *DLNARenderer) Status(ctx context.Context) (Status, error) {
	st := Status{Volume: -1}
	info, err := d.soap(ctx, d.transportURL, d.transportType, "GetTransportInfo", "<InstanceID>0</InstanceID>")
	if err != nil {
		return st, err
	}
	switch tag(info, "CurrentTransportState") {
	case "PLAYING":
		st.State = StatePlaying
	case "PAUSED_PLAYBACK", "PAUSED_RECORDING":
		st.State = StatePaused
	case "TRANSITIONING":
		st.State = StateBuffering
	case "STOPPED":
		st.State = StateStopped
	default:
		st.State = StateIdle
	}
	if pos, err := d.soap(ctx, d.transportURL, d.transportType, "GetPositionInfo", "<InstanceID>0</InstanceID>"); err == nil {
		st.Position = parseHMS(tag(pos, "RelTime"))
		st.Duration = parseHMS(tag(pos, "TrackDuration"))
	}
	if st.Duration == 0 {
		d.mu.Lock()
		st.Duration = d.duration
		d.mu.Unlock()
	}
	if d.renderingURL != "" {
		if vol, err := d.soap(ctx, d.renderingURL, d.renderingType, "GetVolume",
			"<InstanceID>0</InstanceID><Channel>Master</Channel>"); err == nil {
			if v, err := strconv.Atoi(tag(vol, "CurrentVolume")); err == nil {
				st.Volume = float64(v) / 100
			}
		}
	}
	return st, nil
}

func (d *DLNARenderer) Close() error { return nil }

// soap calls action on one of the renderer's services.
func (d *DLNARenderer) soap(ctx context.Context, controlURL, service, action, args string) (string, error) {
	return soapCall(ctx, d.name, controlURL, service, action, args)
}
