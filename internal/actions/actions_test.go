package actions

import (
	"testing"

	"gioui.org/io/key"
)

func TestSame(t *testing.T) {
	for _, c := range [][2]string{
		{"mod+shift+o", "Mod+Shift+O"},
		{"Shift+Mod+O", "Mod+Shift+O"},
		{"space", "Space"},
		{"Alt+return", "Alt+Enter"},
		{"shift+/", "Shift+/"},
	} {
		if !Same(c[0], c[1]) {
			t.Errorf("Same(%q, %q) = false", c[0], c[1])
		}
	}
	if Same("Mod+O", "O") {
		t.Error("Same(Mod+O, O) = true")
	}
}

func TestFromEvent(t *testing.T) {
	for _, c := range []struct {
		e    key.Event
		want string
	}{
		{key.Event{Name: "O", Modifiers: key.ModShortcut | key.ModShift}, "Mod+Shift+O"},
		{key.Event{Name: key.NameLeftArrow, Modifiers: key.ModAlt}, "Alt+Left"},
		{key.Event{Name: key.NameEnter}, "Enter"},
		{key.Event{Name: key.NameF11}, "F11"},
	} {
		got := FromEvent(c.e)
		if got != c.want {
			t.Errorf("FromEvent(%v) = %q, want %q", c.e, got, c.want)
		}
		// What the shortcut editor records must run the command again.
		r := NewRegistry()
		r.Add(&Action{ID: "x", Keys: []string{got}})
		if a := r.Lookup(c.e); a == nil {
			t.Errorf("%q doesn't match the press it came from", got)
		}
	}
}

func TestLookup(t *testing.T) {
	r := NewRegistry()
	ran := ""
	r.Add(
		&Action{ID: "open", Title: "Open File", Category: "File", Keys: []string{"Mod+O"}, Run: func() { ran = "open" }},
		&Action{ID: "pause", Title: "Play or Pause", Category: "Playback", Keys: []string{"Space", "K"}, Run: func() { ran = "pause" }},
	)
	if a := r.Lookup(key.Event{Name: "O", Modifiers: key.ModShortcut}); a == nil || a.ID != "open" {
		t.Fatalf("Mod+O = %v", a)
	}
	if a := r.Lookup(key.Event{Name: key.NameSpace}); a == nil || a.ID != "pause" {
		t.Fatalf("Space = %v", a)
	}
	if a := r.Lookup(key.Event{Name: "O"}); a != nil {
		t.Fatalf("plain O = %v", a.ID)
	}

	// Rebinding moves the shortcut.
	r.SetCustom("open", []string{"F2"})
	if a := r.Lookup(key.Event{Name: "O", Modifiers: key.ModShortcut}); a != nil {
		t.Error("old binding still active")
	}
	if a := r.Lookup(key.Event{Name: key.NameF2}); a == nil || a.ID != "open" {
		t.Error("new binding missing")
	}
	r.SetCustom("open", nil)
	if !r.Run(r.Lookup(key.Event{Name: "O", Modifiers: key.ModShortcut}).ID) || ran != "open" {
		t.Error("default binding not restored")
	}
}

func TestSearch(t *testing.T) {
	r := NewRegistry()
	noop := func() {}
	r.Add(
		&Action{ID: "a", Title: "Toggle Fullscreen", Category: "View", Run: noop},
		&Action{ID: "b", Title: "Full Speed Ahead", Category: "Playback", Run: noop},
		&Action{ID: "c", Title: "Hidden", Category: "View", Run: noop, Hidden: true},
	)
	got := r.Search("full")
	if len(got) != 2 || got[0].ID != "b" {
		t.Errorf("search full = %v", ids(got))
	}
	if got := r.Search("view toggle"); len(got) != 1 || got[0].ID != "a" {
		t.Errorf("search view toggle = %v", ids(got))
	}
}

func ids(as []*Action) []string {
	var out []string
	for _, a := range as {
		out = append(out, a.ID)
	}
	return out
}

func TestMPVKey(t *testing.T) {
	for _, c := range []struct {
		e    key.Event
		want string
	}{
		{key.Event{Name: "A"}, "a"},
		{key.Event{Name: "A", Modifiers: key.ModShift}, "A"},
		{key.Event{Name: key.NameLeftArrow, Modifiers: key.ModCtrl}, "ctrl+LEFT"},
		{key.Event{Name: key.NameSpace}, "SPACE"},
		{key.Event{Name: "1", Modifiers: key.ModAlt}, "alt+1"},
	} {
		if got := MPVKey(c.e); got != c.want {
			t.Errorf("MPVKey(%v) = %q, want %q", c.e, got, c.want)
		}
	}
}
