package cast

import (
	"encoding/xml"
	"strconv"
	"strings"
)

// didlLite is the DIDL-Lite listing in a Browse response.
type didlLite struct {
	Containers []struct {
		ID         string `xml:"id,attr"`
		ChildCount string `xml:"childCount,attr"`
		Title      string `xml:"title"`
	} `xml:"container"`
	Items []struct {
		ID     string `xml:"id,attr"`
		Title  string `xml:"title"`
		Class  string `xml:"class"`
		Artist string `xml:"artist"`
		Album  string `xml:"album"`
		Res    []struct {
			ProtocolInfo string `xml:"protocolInfo,attr"`
			Duration     string `xml:"duration,attr"`
			URL          string `xml:",chardata"`
		} `xml:"res"`
	} `xml:"item"`
}

// parseDIDL reads a DIDL-Lite document into entries, in its order:
// containers, then items.
func parseDIDL(doc string) ([]MediaEntry, error) {
	if strings.TrimSpace(doc) == "" {
		return nil, nil
	}
	var d didlLite
	if err := xml.Unmarshal([]byte(doc), &d); err != nil {
		return nil, err
	}
	var out []MediaEntry
	for _, c := range d.Containers {
		count := -1
		if n, err := strconv.Atoi(c.ChildCount); err == nil {
			count = n
		}
		out = append(out, MediaEntry{ID: c.ID, Title: strings.TrimSpace(c.Title), Folder: true, Count: count})
	}
	for _, it := range d.Items {
		e := MediaEntry{ID: it.ID, Title: strings.TrimSpace(it.Title), Artist: it.Artist, Album: it.Album}
		class := strings.ToLower(it.Class)
		e.Video = strings.HasPrefix(class, "object.item.videoitem")
		e.Audio = strings.HasPrefix(class, "object.item.audioitem")
		// The first resource is the original; later ones are usually
		// transcodes or thumbnails.
		for _, r := range it.Res {
			mime := protocolMIME(r.ProtocolInfo)
			if !e.Video && !e.Audio {
				e.Video = strings.HasPrefix(mime, "video/")
				e.Audio = strings.HasPrefix(mime, "audio/")
			}
			if (e.Video && strings.HasPrefix(mime, "video/")) || (e.Audio && strings.HasPrefix(mime, "audio/")) || mime == "" {
				e.URL = strings.TrimSpace(r.URL)
				e.Duration = parseHMS(r.Duration)
				break
			}
		}
		out = append(out, e)
	}
	return out, nil
}

// protocolMIME is the MIME type in a res protocolInfo
// ("http-get:*:video/mp4:DLNA.ORG_PN=...").
func protocolMIME(info string) string {
	parts := strings.Split(info, ":")
	if len(parts) < 3 {
		return ""
	}
	return strings.ToLower(parts[2])
}
