package settings

import "strings"

// ParseOptions reads "name=value" lines (blank lines and # comments
// ignored; a bare name means "yes").
func ParseOptions(text string) []MPVOption {
	var out []MPVOption
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "--")
		name, value, ok := strings.Cut(line, "=")
		if !ok {
			value = "yes"
		}
		out = append(out, MPVOption{strings.TrimSpace(name), strings.TrimSpace(value)})
	}
	return out
}
