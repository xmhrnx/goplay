// Package actions is goplay's command registry: every command the player
// offers, with its title, category and default keyboard shortcuts. The
// menus, the keyboard, the command palette and the shortcut editor are
// all built from it, so a command added here shows up everywhere.
package actions

import (
	"sort"
	"strings"

	"gioui.org/io/key"
)

// Registry holds the actions and the effective key bindings.
type Registry struct {
	list []*Action
	byID map[string]*Action
	// keys maps a key and its modifiers to the action they run.
	keys map[combo]*Action
	// custom holds user overrides: action ID to its shortcuts.
	custom map[string][]string
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{byID: map[string]*Action{}, keys: map[combo]*Action{}, custom: map[string][]string{}}
}

// Add registers actions. IDs must be unique.
func (r *Registry) Add(as ...*Action) {
	for _, a := range as {
		if _, dup := r.byID[a.ID]; dup {
			panic("actions: duplicate id " + a.ID)
		}
		r.list = append(r.list, a)
		r.byID[a.ID] = a
	}
	r.rebind()
}

// Get returns an action by ID, or nil.
func (r *Registry) Get(id string) *Action { return r.byID[id] }

// Run runs an action by ID if it exists and is enabled.
func (r *Registry) Run(id string) bool {
	a := r.byID[id]
	if a == nil || !a.IsEnabled() || a.Run == nil {
		return false
	}
	a.Run()
	return true
}

// All returns the actions in registration order.
func (r *Registry) All() []*Action { return r.list }

// Categories returns the categories in first-seen order.
func (r *Registry) Categories() []string {
	var out []string
	seen := map[string]bool{}
	for _, a := range r.list {
		if !seen[a.Category] {
			seen[a.Category] = true
			out = append(out, a.Category)
		}
	}
	return out
}

// SetCustom replaces an action's shortcuts (nil restores the defaults).
func (r *Registry) SetCustom(id string, keys []string) {
	if keys == nil {
		delete(r.custom, id)
	} else {
		r.custom[id] = keys
	}
	r.rebind()
}

// Keys returns an action's effective shortcuts.
func (r *Registry) Keys(id string) []string {
	if k, ok := r.custom[id]; ok {
		return k
	}
	if a := r.byID[id]; a != nil {
		return a.Keys
	}
	return nil
}

// Customized reports whether the user changed an action's shortcuts.
func (r *Registry) Customized(id string) bool {
	_, ok := r.custom[id]
	return ok
}

// Owner returns the action a shortcut runs, if any.
func (r *Registry) Owner(spec string) *Action {
	c, ok := parse(spec)
	if !ok {
		return nil
	}
	return r.keys[c]
}

// rebind rebuilds the shortcut index. Later registrations and custom
// bindings win over earlier defaults.
func (r *Registry) rebind() {
	r.keys = map[combo]*Action{}
	bind := func(a *Action, specs []string) {
		for _, k := range specs {
			if c, ok := parse(k); ok {
				r.keys[c] = a
			}
		}
	}
	for _, a := range r.list {
		if _, ok := r.custom[a.ID]; !ok {
			bind(a, a.Keys)
		}
	}
	for id, ks := range r.custom {
		if a := r.byID[id]; a != nil {
			bind(a, ks)
		}
	}
}

// Lookup returns the action bound to a key press.
func (r *Registry) Lookup(e key.Event) *Action {
	name := e.Name
	if name == key.NameEnter {
		name = key.NameReturn // the keypad's Enter runs Enter's command
	}
	return r.keys[combo{name, e.Modifiers}]
}

// Search returns the palette's matches for a query: every word must
// appear in the title or category; titles that start with the query
// come first.
func (r *Registry) Search(query string) []*Action {
	q := strings.ToLower(strings.TrimSpace(query))
	words := strings.Fields(q)
	var out []*Action
	for _, a := range r.list {
		if a.Hidden || a.Run == nil {
			continue
		}
		hay := strings.ToLower(a.Title + " " + a.Category)
		ok := true
		for _, w := range words {
			if !strings.Contains(hay, w) {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, a)
		}
	}
	if q != "" {
		sort.SliceStable(out, func(i, j int) bool {
			pi := strings.HasPrefix(strings.ToLower(out[i].Title), q)
			pj := strings.HasPrefix(strings.ToLower(out[j].Title), q)
			return pi && !pj
		})
	}
	return out
}
