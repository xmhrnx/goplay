package library

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Discs. mpv plays a DVD or Blu-ray from its folder structure, in a
// drive, copied to a folder or in a disc image, through a dvd:// or
// bd:// address naming the device; a disc opened as a folder of files
// would be a pile of numbered video files instead of the movie.

// DiscKind is the kind of disc at a path.
type DiscKind int

const (
	NoDisc DiscKind = iota
	DVD
	BluRay
)

// Disc tells whether path is a DVD or Blu-ray: a disc's root folder
// (in a drive or copied), its VIDEO_TS or BDMV folder, or an .iso
// image. root is what mpv opens: the disc's root, or the image.
func Disc(path string) (kind DiscKind, root string) {
	fi, err := os.Stat(path)
	if err != nil {
		return NoDisc, ""
	}
	if !fi.IsDir() {
		if strings.EqualFold(filepath.Ext(path), ".iso") {
			return imageKind(path, fi.Size()), path
		}
		return NoDisc, ""
	}
	switch base := strings.ToUpper(filepath.Base(path)); {
	case base == "VIDEO_TS" && hasEntry(path, "VIDEO_TS.IFO"):
		return DVD, filepath.Dir(path)
	case base == "BDMV" && hasEntry(path, "index.bdmv"):
		return BluRay, filepath.Dir(path)
	}
	if d := entry(path, "VIDEO_TS"); d != "" && hasEntry(d, "VIDEO_TS.IFO") {
		return DVD, path
	}
	if d := entry(path, "BDMV"); d != "" && hasEntry(d, "index.bdmv") {
		return BluRay, path
	}
	return NoDisc, ""
}

// DiscURL is the address mpv plays the disc at root with. A Blu-ray
// starts at its longest title, the movie; so does a DVD.
func DiscURL(kind DiscKind, root string) string {
	if kind == BluRay {
		return "bd://longest/" + root
	}
	return "dvd://longest/" + root
}

// IsDiscURL reports whether u is a disc address made by DiscURL.
func IsDiscURL(u string) bool {
	return strings.HasPrefix(u, "bd://") || strings.HasPrefix(u, "dvd://")
}

// DiscName is the name to show for a disc address: its folder's or
// image's name, or the drive's.
func DiscName(u string) string {
	_, rest, _ := strings.Cut(u, "://")
	_, dev, _ := strings.Cut(rest, "/")
	dev = strings.TrimRight(dev, `/\`)
	if dev == "" {
		return u
	}
	if runtime.GOOS == "windows" && len(dev) == 2 && dev[1] == ':' {
		return dev
	}
	return filepath.Base(dev)
}

// Discs lists the discs in the computer's drives (or mounted images):
// the roots of the mounted volumes holding a DVD or Blu-ray.
func Discs() []string {
	var roots []string
	switch runtime.GOOS {
	case "windows":
		for c := 'A'; c <= 'Z'; c++ {
			roots = append(roots, string(c)+`:\`)
		}
	case "darwin":
		roots = subdirs("/Volumes")
	default:
		user := os.Getenv("USER")
		for _, d := range []string{"/media/" + user, "/run/media/" + user, "/media", "/mnt"} {
			roots = append(roots, subdirs(d)...)
		}
	}
	var discs []string
	for _, r := range roots {
		if k, root := Disc(r); k != NoDisc && root == r {
			discs = append(discs, r)
		}
	}
	return discs
}

func subdirs(dir string) []string {
	entries, _ := os.ReadDir(dir)
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	return out
}

// entry finds name in dir ignoring case (discs copied or mounted on some
// systems show their names in lower case), returning its path.
func entry(dir, name string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if strings.EqualFold(e.Name(), name) {
			return filepath.Join(dir, e.Name())
		}
	}
	return ""
}

func hasEntry(dir, name string) bool { return entry(dir, name) != "" }

// imageKind tells a DVD image from a Blu-ray one by the folder names in
// its file system, which lies at the start of the image. Without
// either, the size decides: a DVD holds at most 8.5 GB.
func imageKind(path string, size int64) DiscKind {
	f, err := os.Open(path)
	if err != nil {
		return NoDisc
	}
	defer f.Close()
	buf, _ := io.ReadAll(io.LimitReader(f, 32<<20))
	switch {
	case bytes.Contains(buf, []byte("BDMV")):
		return BluRay
	case bytes.Contains(buf, []byte("VIDEO_TS")):
		return DVD
	case size > 8_600_000_000:
		return BluRay
	}
	return DVD
}
