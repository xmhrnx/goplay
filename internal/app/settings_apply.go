package app

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/goplay/internal/player"
	"github.com/mehran9675/goplay/internal/settings"
)

// setSetting stores a setting and puts it into effect.
func (a *App) setSetting(key, value string) {
	if a.settings.Get(key) == value {
		return
	}
	a.settings.Set(key, value)
	a.applySetting(key)
}

// applySetting puts a stored setting into effect: its mpv options, or
// the app behavior it controls.
func (a *App) applySetting(key string) {
	d := settings.Lookup(key)
	if d == nil {
		return
	}
	v := a.settings.Get(key)
	switch key {
	case settings.KeyLanguage:
		giode.SetLocale(uiLocale(v))
		a.rebuildUI()
		a.player.Set("alang", a.audioLanguages())
		return
	case settings.KeyAudioLang, settings.KeyAudioMatchUI:
		a.player.Set("alang", a.audioLanguages())
		return
	case settings.KeyAlwaysOnTop:
		a.setOnTop(v == "yes")
	case settings.KeyAudioDevice:
		// The player follows the device as it comes and goes.
		a.player.SetAudioDevice(v)
		return
	case settings.KeyStereoView:
		if a.stereo.layout != player.Layout2D {
			a.stereo.view = player.StereoView(v)
			a.applyStereo()
		}
	case settings.KeyVRFOV:
		if a.vrActive() {
			a.stereo.vr.FOV = a.settings.Float(settings.KeyVRFOV)
			a.applyStereo()
		}
	}
	if d.Restart {
		a.toast(tr("{setting}: takes effect when GoPlay restarts", "setting", tr(d.Label)))
		return
	}
	var opts []settings.MPVOption
	if d.Group != "" {
		opts = []settings.MPVOption{a.settings.GroupOption(d.Group)}
	} else {
		opts = d.MPVOptions(v)
	}
	for _, o := range opts {
		if err := a.player.Set(o.Name, o.Value); err != nil {
			a.toastError(tr("{setting}: mpv didn't accept {option}", "setting", tr(d.Label), "option", o.Name+"="+o.Value))
		}
	}
}

// audioLanguages is mpv's alang, the languages to prefer for audio: the
// Preferred language, or else the app's own if that setting is on. Where
// a file has none of them, mpv plays its default track.
func (a *App) audioLanguages() string {
	if v := a.settings.Get(settings.KeyAudioLang); v != "" {
		return v
	}
	if a.settings.Bool(settings.KeyAudioMatchUI) {
		return player.LanguageCodes(uiLocale(a.settings.Get(settings.KeyLanguage)))
	}
	return ""
}
