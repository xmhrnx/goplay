package library

import (
	"os"
	"path/filepath"
	"runtime"
)

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
