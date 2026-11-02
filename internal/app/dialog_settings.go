package app

import (
	"github.com/mehran9675/giode/components/button"
	"github.com/mehran9675/giode/components/dialog"
	"github.com/mehran9675/giode/components/scroll"
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/settings"
)

// The settings dialog: a page per section from settings.Defs, each row a
// prefControl, plus the shortcut editor.

const shortcutsPage = "Shortcuts"

type prefsView struct {
	modal
	u        *ui
	page     string
	controls map[string]*prefControl
	scroll   *scroll.Scroll
	keys     *shortcutEditor

	done, reset, folder, mpvConf, inputConf *button.Button
}

func newPrefsView(u *ui) *prefsView {
	a := u.a
	p := &prefsView{u: u, page: settings.Sections[0], controls: map[string]*prefControl{}, keys: newShortcutEditor(u)}
	for i := range settings.Defs {
		d := &settings.Defs[i]
		p.controls[d.Key] = newPrefControl(a, d, d.Options)
	}
	p.scroll = scroll.New(func() elements.Element { return clearOfBar(p.pageContent()) }, scrollStyles())
	p.done = primaryButton(tr("Done"), func() { p.open = false })
	p.reset = secondaryButton(tr("Reset All"), func() {
		a.settings.Reset()
		for i := range settings.Defs {
			a.applySetting(settings.Defs[i].Key)
		}
		a.toast(tr("Settings reset to their defaults"))
	})
	p.folder = secondaryButton(tr("Open Config Folder"), a.openConfigFolder)
	p.mpvConf = secondaryButton(tr("Edit mpv.conf"), func() { a.editConfigFile("mpv.conf") })
	p.inputConf = secondaryButton(tr("Edit input.conf"), func() { a.editConfigFile("input.conf") })
	p.dlg = dialog.New(&p.open, p.content, styles.Styles{
		Background: colSurface, BorderRadius: 16,
		BoxShadow: properties.BoxShadow{Y: 18, Blur: 48, Color: "#000000e6"},
	}).Padding(properties.Inset{})
	return p
}

// show opens the dialog at page, or where it was left for "".
func (p *prefsView) show(page string) {
	if page != "" {
		p.page = page
	}
	p.refresh()
	p.open = true
}

// close cancels recording a shortcut first, then closes.
func (p *prefsView) close() {
	if p.keys.capture != "" {
		p.keys.capture = ""
		return
	}
	p.open = false
}

// refresh reloads what changes while the dialog is closed: the audio
// devices and every value.
func (p *prefsView) refresh() {
	d := settings.Lookup(settings.KeyAudioDevice)
	p.controls[d.Key] = newPrefControl(p.u.a, d, p.u.a.audioDevices())
	p.sync()
}

// sync sets the controls from the stored values, except one in use.
func (p *prefsView) sync() {
	store := p.u.a.settings
	for key, c := range p.controls {
		c.sync(store.Get(key))
	}
}

func (p *prefsView) content() elements.Element {
	p.sync()
	pages := append(append([]string{}, settings.Sections...), shortcutsPage)
	nav := []elements.Element{elements.Box(styles.Styles{Padding: properties.Inset{Left: 10, Bottom: 10}}, heading(tr("Settings")))}
	for _, pg := range pages {
		nav = append(nav, p.navItem(pg))
	}
	sidebar := elements.Box(styles.Styles{Width: 200, Height: -1, FlexDirection: properties.FlexDirectionColumn, Gap: 2,
		Padding: properties.UniformInset(14), Background: "#111113"}, nav...)

	body := elements.Element(p.scroll)
	var footer []elements.Element
	switch p.page {
	case shortcutsPage:
		body = p.keys.scroll
		footer = append(footer, p.keys.resetAll)
	case "Advanced":
		footer = append(footer, p.folder, p.mpvConf, p.inputConf)
	default:
		footer = append(footer, p.reset)
	}
	footer = append(footer, elements.Box(styles.Styles{FlexGrow: 1, Width: -1}), p.done)
	main := elements.Box(styles.Styles{FlexGrow: 1, Width: -1, Height: -1, FlexDirection: properties.FlexDirectionColumn},
		elements.Box(styles.Styles{FlexGrow: 1, Width: -1, Height: -1, Padding: properties.Inset{Top: 18, Left: 24, Right: 12}}, body),
		elements.Row(styles.Styles{Width: -1, Gap: 8, Align: properties.AlignCenter, Wrap: properties.WrapNoWrap,
			Padding: properties.SymmetricInset(20, 14), BorderColor: colBorder}, footer...),
	)
	// As large as comfortable, but never larger than the window.
	u := p.u
	w := min(860, int(float32(u.fb.X)/u.scale)-48)
	h := min(620, int(float32(u.fb.Y)/u.scale)-48)
	return elements.Row(styles.Styles{Width: properties.Width(max(w, 320)), Height: properties.Height(max(h, 300)), Wrap: properties.WrapNoWrap}, sidebar, main)
}

func (p *prefsView) navItem(page string) elements.Element {
	return elements.Keyed("settings-nav", page, func() *elements.Pressable {
		return elements.NewPressable(func(s elements.PressState) elements.Element {
			bg, fg := "", colMuted
			switch {
			case p.page == page:
				bg, fg = colSurface3, colText
			case s.Hovered:
				bg, fg = colSurface2, colText
			}
			return elements.Box(styles.Styles{Width: -1, Background: bg, BorderRadius: 8, Padding: properties.SymmetricInset(10, 8)},
				elements.Text(tr(page), styles.Styles{Color: fg, FontSize: 13, FontWeight: properties.FontWeightMedium}))
		}).Label(tr(page)).OnClick(func(elements.ClickEvent) {
			p.page = page
			p.keys.capture = ""
		})
	})
}

// pageContent is the current section's settings.
func (p *prefsView) pageContent() elements.Element {
	var rows []elements.Element
	for i := range settings.Defs {
		d := &settings.Defs[i]
		if d.Section == p.page {
			rows = append(rows, p.controls[d.Key].row(p.u.a.settings.Get(d.Key)))
		}
	}
	return elements.Box(styles.Styles{Width: -1, FlexDirection: properties.FlexDirectionColumn, Gap: 4, Padding: properties.Inset{Right: 12, Bottom: 12}}, rows...)
}

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
