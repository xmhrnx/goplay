package app

import (
	"github.com/mehran9675/goplay/internal/settings"
)

// audioDevices lists mpv's outputs for the device setting.
func (a *App) audioDevices() []settings.Option {
	out := []settings.Option{{Label: "Automatic", Value: "auto"}}
	var list []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if a.player.Handle().JSON("audio-device-list", &list) != nil {
		return out
	}
	for _, d := range list {
		if d.Name == "" || d.Name == "auto" {
			continue
		}
		label := d.Description
		if label == "" {
			label = d.Name
		}
		out = append(out, settings.Option{Label: label, Value: d.Name})
	}
	return out
}
