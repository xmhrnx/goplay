// Package library keeps the recently played list and finds the media
// files in a folder.
package library

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
)

// Folder lists the media files in dir, naturally sorted.
func Folder(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && IsMedia(e.Name()) && !strings.HasPrefix(e.Name(), ".") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	slices.SortFunc(files, func(a, b string) int {
		switch {
		case NaturalLess(a, b):
			return -1
		case NaturalLess(b, a):
			return 1
		}
		return 0
	})
	if len(files) == 0 {
		return nil, errors.New("no video or audio files in " + filepath.Base(dir))
	}
	return files, nil
}

// Expand turns what the user opened - files, folders, URLs - into a
// playlist, expanding folders to their media files and making DVDs and
// Blu-rays disc addresses (see Disc). With queueFolder, a single file
// opened alone brings the rest of its folder along; start is then its
// index in the list.
func Expand(items []string, queueFolder bool) (list []string, start int) {
	list, start = expand(items, queueFolder)
	for i, it := range list {
		if !IsURL(it) {
			if k, root := Disc(it); k != NoDisc {
				list[i] = DiscURL(k, root)
			}
		}
	}
	return list, start
}

func expand(items []string, queueFolder bool) (list []string, start int) {
	for _, it := range items {
		if IsURL(it) {
			list = append(list, it)
			continue
		}
		if k, _ := Disc(it); k != NoDisc {
			list = append(list, it)
			continue
		}
		if fi, err := os.Stat(it); err == nil && fi.IsDir() {
			if files, err := Folder(it); err == nil {
				list = append(list, files...)
			}
			continue
		}
		list = append(list, it)
	}
	if queueFolder && len(items) == 1 && len(list) == 1 && !IsURL(list[0]) && !isDir(list[0]) {
		if files, err := Folder(filepath.Dir(list[0])); err == nil && len(files) > 1 {
			for i, f := range files {
				if sameFile(f, list[0]) {
					return files, i
				}
			}
		}
	}
	return list, 0
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

func sameFile(a, b string) bool {
	if a == b {
		return true
	}
	fa, err1 := os.Stat(a)
	fb, err2 := os.Stat(b)
	return err1 == nil && err2 == nil && os.SameFile(fa, fb)
}

// NaturalLess orders strings as people expect: digit runs compare by
// value ("ep2" before "ep10"), letters case-insensitively.
func NaturalLess(a, b string) bool {
	ar, br := []rune(a), []rune(b)
	i, j := 0, 0
	for i < len(ar) && j < len(br) {
		if unicode.IsDigit(ar[i]) && unicode.IsDigit(br[j]) {
			si, sj := i, j
			for i < len(ar) && unicode.IsDigit(ar[i]) {
				i++
			}
			for j < len(br) && unicode.IsDigit(br[j]) {
				j++
			}
			na := strings.TrimLeft(string(ar[si:i]), "0")
			nb := strings.TrimLeft(string(br[sj:j]), "0")
			if len(na) != len(nb) {
				return len(na) < len(nb)
			}
			if na != nb {
				return na < nb
			}
			continue
		}
		ca, cb := unicode.ToLower(ar[i]), unicode.ToLower(br[j])
		if ca != cb {
			return ca < cb
		}
		i++
		j++
	}
	return len(ar)-i < len(br)-j
}
