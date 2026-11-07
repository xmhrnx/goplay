package library

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

// History is the recently played list, saved as JSON. It is safe for
// concurrent use.
type History struct {
	mu    sync.Mutex
	path  string
	items []Recent
}

// maxRecent is how many entries the list keeps.
const maxRecent = 30

// LoadHistory reads the list saved at path; a missing file is empty.
func LoadHistory(path string) (*History, error) {
	h := &History{path: path}
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return h, nil
	}
	if err != nil {
		return h, err
	}
	return h, json.Unmarshal(bytes.TrimPrefix(data, utf8BOM), &h.items)
}

// Items returns the list, most recent first.
func (h *History) Items() []Recent {
	h.mu.Lock()
	defer h.mu.Unlock()
	return slices.Clone(h.items)
}

// Played moves path to the top of the list, updating its details.
func (h *History) Played(r Recent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r.Played.IsZero() {
		r.Played = time.Now()
	}
	h.items = slices.DeleteFunc(h.items, func(x Recent) bool { return x.Path == r.Path })
	h.items = slices.Insert(h.items, 0, r)
	if len(h.items) > maxRecent {
		h.items = h.items[:maxRecent]
	}
}

// Update records the position reached in path, without reordering.
func (h *History) Update(path, title string, position, duration float64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := range h.items {
		if h.items[i].Path == path {
			if title != "" {
				h.items[i].Title = title
			}
			h.items[i].Position, h.items[i].Duration = position, duration
			return
		}
	}
}

// Remove drops one entry; Clear drops them all.
func (h *History) Remove(path string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.items = slices.DeleteFunc(h.items, func(x Recent) bool { return x.Path == path })
}

func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.items = nil
}

// Save writes the list.
func (h *History) Save() error {
	if h.path == "" {
		return nil
	}
	h.mu.Lock()
	data, err := json.MarshalIndent(h.items, "", "  ")
	h.mu.Unlock()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(h.path), 0o755); err != nil {
		return err
	}
	tmp := h.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, h.path)
}

// utf8BOM is the byte-order mark some editors (Windows Notepad) write
// at the start of UTF-8 files.
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}
