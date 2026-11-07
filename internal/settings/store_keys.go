package settings

import "strings"

// Keys returns the user's key bindings for an action, or nil when they
// haven't changed the defaults. Bindings are stored as "key.<action>",
// a comma-separated list of shortcuts ("Ctrl+O, F1"); "none" unbinds.
func (s *Store) Keys(action string) ([]string, bool) {
	s.mu.Lock()
	v, ok := s.values["key."+action]
	s.mu.Unlock()
	if !ok {
		return nil, false
	}
	if v == "none" {
		return []string{}, true
	}
	var out []string
	for _, k := range strings.Split(v, ",") {
		if k = strings.TrimSpace(k); k != "" {
			out = append(out, k)
		}
	}
	return out, true
}

// SetKeys stores the bindings for an action; nil restores the default.
func (s *Store) SetKeys(action string, keys []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := "key." + action
	switch {
	case keys == nil:
		delete(s.values, k)
	case len(keys) == 0:
		s.values[k] = "none"
	default:
		s.values[k] = strings.Join(keys, ", ")
	}
	s.dirty = true
}

// ResetKeys restores every default key binding.
func (s *Store) ResetKeys() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.values {
		if strings.HasPrefix(k, "key.") {
			delete(s.values, k)
			s.dirty = true
		}
	}
}
