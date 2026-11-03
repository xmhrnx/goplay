package library

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDisc(t *testing.T) {
	dir := t.TempDir()
	dvd := filepath.Join(dir, "Movie DVD")
	bd := filepath.Join(dir, "Movie BD")
	os.MkdirAll(filepath.Join(dvd, "VIDEO_TS"), 0o755)
	os.WriteFile(filepath.Join(dvd, "VIDEO_TS", "VIDEO_TS.IFO"), nil, 0o644)
	os.WriteFile(filepath.Join(dvd, "VIDEO_TS", "VTS_01_1.VOB"), nil, 0o644)
	// Copied discs sometimes have their names in lower case.
	os.MkdirAll(filepath.Join(bd, "bdmv"), 0o755)
	os.WriteFile(filepath.Join(bd, "bdmv", "index.bdmv"), nil, 0o644)
	os.WriteFile(filepath.Join(dir, "dvd.iso"), append(make([]byte, 40000), "VIDEO_TS"...), 0o644)
	os.WriteFile(filepath.Join(dir, "bd.iso"), append(make([]byte, 40000), "\x08BDMV"...), 0o644)
	os.MkdirAll(filepath.Join(dir, "Plain", "VIDEO_TS"), 0o755)

	for _, c := range []struct {
		path string
		kind DiscKind
		root string
	}{
		{dvd, DVD, dvd},
		{filepath.Join(dvd, "VIDEO_TS"), DVD, dvd},
		{bd, BluRay, bd},
		{filepath.Join(bd, "bdmv"), BluRay, bd},
		{filepath.Join(dir, "dvd.iso"), DVD, filepath.Join(dir, "dvd.iso")},
		{filepath.Join(dir, "bd.iso"), BluRay, filepath.Join(dir, "bd.iso")},
		// A VIDEO_TS folder without its IFO isn't a disc.
		{filepath.Join(dir, "Plain"), NoDisc, ""},
		{filepath.Join(dvd, "VIDEO_TS", "VTS_01_1.VOB"), NoDisc, ""},
	} {
		if k, root := Disc(c.path); k != c.kind || root != c.root {
			t.Errorf("Disc(%s) = %v %q, want %v %q", c.path, k, root, c.kind, c.root)
		}
	}

	// Opened, discs become disc addresses; the folder isn't played as
	// a pile of VOB files.
	list, _ := Expand([]string{filepath.Join(dvd, "VIDEO_TS"), bd, filepath.Join(dir, "bd.iso")}, true)
	want := []string{"dvd://longest/" + dvd, "bd://longest/" + bd, "bd://longest/" + filepath.Join(dir, "bd.iso")}
	if len(list) != len(want) {
		t.Fatalf("Expand = %v", list)
	}
	for i := range want {
		if list[i] != want[i] {
			t.Errorf("Expand[%d] = %q, want %q", i, list[i], want[i])
		}
	}
	if n := (Recent{Path: list[0]}).Name(); n != "Movie DVD" {
		t.Errorf("name = %q", n)
	}
	// An image opened alone still queues the rest of its folder.
	list, start := Expand([]string{filepath.Join(dir, "dvd.iso")}, true)
	if len(list) != 2 || list[start] != "dvd://longest/"+filepath.Join(dir, "dvd.iso") {
		t.Errorf("image in its folder: %v start %d", list, start)
	}
}

func TestParseDiscTitle(t *testing.T) {
	got, ok := ParseDiscTitle("title: 91 (00:05:33.333) (00965.mpls)")
	if !ok || got.Number != 91 || got.Length != 5*time.Minute+33333*time.Millisecond {
		t.Errorf("got %+v %v", got, ok)
	}
	if got, ok := ParseDiscTitle("title: 1 (01:56:02.289) (00040.mpls)"); !ok || got.Length < 116*time.Minute {
		t.Errorf("movie: %+v %v", got, ok)
	}
	if _, ok := ParseDiscTitle("Director's Cut"); ok {
		t.Error("a file's edition parsed as a disc title")
	}
}
