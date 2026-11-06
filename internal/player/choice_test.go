package player

import "testing"

func TestTrackChoice(t *testing.T) {
	ep1 := []Track{
		{ID: 1, Kind: Video},
		{ID: 1, Kind: Audio, Lang: "jpn", Default: true},
		{ID: 2, Kind: Audio, Lang: "eng"},
		{ID: 1, Kind: Subtitle, Lang: "eng", Title: "Signs", Forced: true},
		{ID: 2, Kind: Subtitle, Lang: "eng", Title: "Full"},
	}
	// The next episode orders its tracks differently.
	ep2 := []Track{
		{ID: 1, Kind: Audio, Lang: "eng"},
		{ID: 2, Kind: Audio, Lang: "jpn", Default: true},
		{ID: 1, Kind: Subtitle, Lang: "eng", Title: "Full"},
		{ID: 2, Kind: Subtitle, Lang: "eng", Title: "Signs", Forced: true},
		{ID: 3, Kind: Subtitle, Lang: "spa"},
	}
	tests := []struct {
		name   string
		k      TrackKind
		id     int
		want   int
		wantOK bool
	}{
		{"audio by language", Audio, 2, 1, true},
		{"subtitle by title", Subtitle, 2, 1, true},
		{"forced subtitle", Subtitle, 1, 2, true},
		{"subtitles off", Subtitle, 0, 0, true},
	}
	for _, tt := range tests {
		c := ChoiceOf(ep1, tt.k, tt.id)
		got, ok := c.Match(ep2, tt.k)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("%s: Match = %d, %v; want %d, %v", tt.name, got, ok, tt.want, tt.wantOK)
		}
	}

	// A language the file lacks leaves mpv's pick alone.
	if _, ok := (TrackChoice{Lang: "fre"}).Match(ep2, Audio); ok {
		t.Error("matched a language the file lacks")
	}
	// Untagged tracks match by position.
	bare := []Track{{ID: 1, Kind: Audio}, {ID: 2, Kind: Audio}}
	if got, ok := ChoiceOf(bare, Audio, 2).Match(bare, Audio); got != 2 || !ok {
		t.Errorf("untagged: Match = %d, %v; want 2, true", got, ok)
	}
}
