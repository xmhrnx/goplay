package cast

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"
)

func describeServer(ctx context.Context, location string) (*MediaServer, error) {
	desc, base, err := describe(ctx, location)
	if err != nil {
		return nil, err
	}
	s := &MediaServer{
		ID:    "upnp:" + desc.Device.UDN,
		Name:  desc.Device.FriendlyName,
		Model: strings.TrimSpace(desc.Device.Manufacturer + " " + desc.Device.ModelName),
	}
	for _, svc := range desc.Device.Services {
		if !strings.HasPrefix(svc.Type, contentDirPrefix) {
			continue
		}
		if ref, err := base.Parse(svc.ControlURL); err == nil {
			s.browseURL, s.browseType = ref.String(), svc.Type
		}
	}
	if s.browseURL == "" {
		return nil, fmt.Errorf("cast: %s has no ContentDirectory service", s.Name)
	}
	if s.Name == "" {
		s.Name = base.Hostname()
	}
	return s, nil
}

// RootID is the ID of a media server's top folder.
const RootID = "0"

// Browse lists the folder id (RootID for the top): its folders first,
// then the video and audio files, each in the server's order. Pictures
// and other files are left out.
func (s *MediaServer) Browse(ctx context.Context, id string) ([]MediaEntry, error) {
	var folders, files []MediaEntry
	for start := 0; start < maxEntries; {
		args := "<ObjectID>" + html.EscapeString(id) + "</ObjectID>" +
			"<BrowseFlag>BrowseDirectChildren</BrowseFlag><Filter>*</Filter>" +
			"<StartingIndex>" + strconv.Itoa(start) + "</StartingIndex>" +
			"<RequestedCount>" + strconv.Itoa(browsePage) + "</RequestedCount><SortCriteria></SortCriteria>"
		resp, err := soapCall(ctx, s.Name, s.browseURL, s.browseType, "Browse", args)
		if err != nil {
			return nil, err
		}
		returned, _ := strconv.Atoi(tag(resp, "NumberReturned"))
		total, _ := strconv.Atoi(tag(resp, "TotalMatches"))
		entries, err := parseDIDL(tag(resp, "Result"))
		if err != nil {
			return nil, fmt.Errorf("cast: %s sent a listing GoPlay can't read: %w", s.Name, err)
		}
		for _, e := range entries {
			switch {
			case e.Folder:
				folders = append(folders, e)
			case e.URL != "" && (e.Video || e.Audio):
				files = append(files, e)
			}
		}
		start += returned
		// Some servers report TotalMatches 0 when they don't count.
		if returned == 0 || (total > 0 && start >= total) || (total == 0 && returned < browsePage) {
			break
		}
	}
	return append(folders, files...), nil
}
