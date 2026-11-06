package player

import "strings"

// TrackChoice is a track the viewer picked, described so the same one
// can be found in other files (the next episode), where its number may
// differ.
type TrackChoice struct {
	Off    bool   `json:"off,omitempty"`
	Lang   string `json:"lang,omitempty"`
	Title  string `json:"title,omitempty"`
	Forced bool   `json:"forced,omitempty"`
	// Index is the track's position among its kind, for files whose
	// tracks have neither language nor title.
	Index int `json:"index"`
}

// ChoiceOf describes track id of kind k in tracks; id 0 is "off".
func ChoiceOf(tracks []Track, k TrackKind, id int) TrackChoice {
	if id == 0 {
		return TrackChoice{Off: true}
	}
	i := 0
	for _, t := range tracks {
		if t.Kind != k {
			continue
		}
		if t.ID == id {
			return TrackChoice{Lang: t.Lang, Title: t.Title, Forced: t.Forced, Index: i}
		}
		i++
	}
	return TrackChoice{Off: true}
}

// Match finds the track of kind k in tracks that c describes: 0 for
// "off", ok false when none fits. The language must agree; the title
// and forced flag pick between tracks in that language.
func (c TrackChoice) Match(tracks []Track, k TrackKind) (id int, ok bool) {
	if c.Off {
		return 0, true
	}
	var same []Track
	for _, t := range tracks {
		if t.Kind == k {
			same = append(same, t)
		}
	}
	if c.Lang == "" && c.Title == "" {
		if c.Index < len(same) && same[c.Index].Lang == "" && same[c.Index].Title == "" {
			return same[c.Index].ID, true
		}
		return 0, false
	}
	best, score := 0, -1
	for _, t := range same {
		if !strings.EqualFold(t.Lang, c.Lang) {
			continue
		}
		s := 0
		if c.Title != "" && strings.EqualFold(t.Title, c.Title) {
			s += 2
		}
		if t.Forced == c.Forced {
			s++
		}
		if s > score {
			best, score = t.ID, s
		}
	}
	return best, score >= 0
}
