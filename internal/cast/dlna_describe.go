package cast

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type upnpDescription struct {
	URLBase string `xml:"URLBase"`
	Device  struct {
		FriendlyName string `xml:"friendlyName"`
		ModelName    string `xml:"modelName"`
		Manufacturer string `xml:"manufacturer"`
		UDN          string `xml:"UDN"`
		Services     []struct {
			Type       string `xml:"serviceType"`
			ControlURL string `xml:"controlURL"`
		} `xml:"serviceList>service"`
	} `xml:"device"`
}

// describe fetches a UPnP device's description and the base its
// relative URLs resolve against.
func describe(ctx context.Context, location string) (*upnpDescription, *url.URL, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	var desc upnpDescription
	if err := xml.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&desc); err != nil {
		return nil, nil, err
	}
	base, err := url.Parse(location)
	if err != nil {
		return nil, nil, err
	}
	if desc.URLBase != "" {
		if b, err := url.Parse(desc.URLBase); err == nil {
			base = b
		}
	}
	return &desc, base, nil
}

// describeRenderer fetches a renderer's description and finds its
// control URLs.
func describeRenderer(ctx context.Context, location string) (*DLNARenderer, error) {
	desc, base, err := describe(ctx, location)
	if err != nil {
		return nil, err
	}
	r := &DLNARenderer{
		name:  desc.Device.FriendlyName,
		model: strings.TrimSpace(desc.Device.Manufacturer + " " + desc.Device.ModelName),
		id:    "dlna:" + desc.Device.UDN,
		host:  base.Host,
	}
	for _, s := range desc.Device.Services {
		ref, err := url.Parse(s.ControlURL)
		if err != nil {
			continue
		}
		u := base.ResolveReference(ref).String()
		switch {
		case strings.HasPrefix(s.Type, "urn:schemas-upnp-org:service:AVTransport:"):
			r.transportURL, r.transportType = u, s.Type
		case strings.HasPrefix(s.Type, "urn:schemas-upnp-org:service:RenderingControl:"):
			r.renderingURL, r.renderingType = u, s.Type
		}
	}
	if r.transportURL == "" {
		return nil, fmt.Errorf("cast: %s has no AVTransport service", r.name)
	}
	if r.name == "" {
		r.name = base.Hostname()
	}
	return r, nil
}
