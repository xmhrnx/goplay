package app

import (
	"math"
	"strconv"

	"github.com/mehran9675/giode/components/dropdown"
	"github.com/mehran9675/giode/components/input"
	"github.com/mehran9675/giode/components/slider"
	"github.com/mehran9675/giode/components/toggle"
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/settings"
)

// prefControl is one setting's row in the settings dialog: its label and
// help, and the switch, list, slider or field that changes it.
type prefControl struct {
	def  *settings.Def
	sw   *toggle.Switch
	sel  *dropdown.Select
	opts []settings.Option
	sl   *slider.Slider
	txt  *input.Input
}

func newPrefControl(a *App, d *settings.Def, opts []settings.Option) *prefControl {
	c := &prefControl{def: d, opts: opts}
	switch d.Kind {
	case settings.Toggle:
		c.sw = toggle.New("", styles.Styles{Background: colSwitchOn, BorderColor: colSurface3}).OnChange(func(on bool) {
			v := "no"
			if on {
				v = "yes"
			}
			a.setSetting(d.Key, v)
		})
	case settings.Choice:
		labels := make([]string, len(opts))
		for i, o := range opts {
			labels[i] = o.Label
			// Languages keep their own names; only Automatic is translated.
			if d.Key != settings.KeyLanguage || o.Value == "" {
				labels[i] = tr(o.Label)
			}
		}
		c.sel = dropdown.New(styles.Styles{Background: colSurface2, Color: colText, BorderColor: colBorder,
			BorderRadius: 8, Width: 250, FontSize: 13}, labels...).HoverColor(colSurface3).OnChange(func(i int) {
			if i >= 0 && i < len(c.opts) {
				a.setSetting(d.Key, c.opts[i].Value)
			}
		})
	case settings.Range:
		c.sl = slider.New(styles.Styles{Color: colHighlight, Background: colTrack, BorderColor: colHighlight}).
			OnChange(func(v float32) { a.setSetting(d.Key, formatRange(d, rangeValue(d, v))) })
	case settings.Text:
		c.txt = input.New("", fieldStyles()).Styles(styles.Styles{Width: 250, Background: colSurface2, Color: colText,
			BorderColor: colBorder, BorderRadius: 8, FontSize: 13})
		apply := func() { a.setSetting(d.Key, c.txt.Text()) }
		if d.Lines > 1 {
			c.txt.Lines(d.Lines, d.Lines).Styles(styles.Styles{Width: -1, Background: colSurface2, Color: colText,
				BorderColor: colBorder, BorderRadius: 8, FontSize: 12})
			c.txt.SubmitOn(input.SubmitNever)
		} else {
			c.txt.Submit(func(string) { apply() })
		}
		c.txt.OnBlur(apply)
	}
	return c
}

// sync shows the stored value v, unless the control is in use.
func (c *prefControl) sync(v string) {
	switch {
	case c.sw != nil:
		c.sw.SetChecked(v == "yes")
	case c.sel != nil && !c.sel.Opened():
		i := 0
		for j, o := range c.opts {
			if o.Value == v {
				i = j
				break
			}
		}
		c.sel.SetSelected(i)
	case c.sl != nil && !c.sl.Dragging():
		f, _ := strconv.ParseFloat(v, 64)
		c.sl.SetValue(float32((f - c.def.Min) / (c.def.Max - c.def.Min)))
	case c.txt != nil && !c.txt.Focused() && c.txt.Text() != v:
		c.txt.SetText(v)
	}
}

// row lays the setting out; v is its stored value.
func (c *prefControl) row(v string) elements.Element {
	d := c.def
	labels := []elements.Element{elements.Text(tr(d.Label), styles.Styles{Color: colText, FontSize: 14})}
	help := ""
	if d.Help != "" {
		help = tr(d.Help)
	}
	if d.Restart {
		if help != "" {
			help += " "
		}
		help += tr("Takes effect when GoPlay restarts.")
	}
	if help != "" {
		labels = append(labels, elements.Text(help, styles.Styles{Color: colFaint, FontSize: 12}))
	}
	label := elements.Box(styles.Styles{FlexGrow: 1, Width: -1, FlexDirection: properties.FlexDirectionColumn, Gap: 3}, labels...)
	var control elements.Element
	switch {
	case c.sw != nil:
		control = c.sw
	case c.sel != nil:
		control = c.sel
	case c.sl != nil:
		f, _ := strconv.ParseFloat(v, 64)
		control = elements.Row(styles.Styles{Gap: 10, Align: properties.AlignCenter, Wrap: properties.WrapNoWrap},
			elements.Box(styles.Styles{Width: 180, Height: 24, Align: properties.AlignCenter}, c.sl),
			elements.Box(styles.Styles{Width: 60, Justify: properties.JustifyEnd},
				elements.Text(formatRange(d, f)+d.Unit, styles.Styles{Color: colMuted, FontSize: 13})))
	case c.txt != nil && d.Lines > 1:
		return elements.Box(styles.Styles{Width: -1, FlexDirection: properties.FlexDirectionColumn, Gap: 8, Padding: properties.SymmetricInset(0, 10)}, label, c.txt)
	case c.txt != nil:
		control = c.txt
	}
	return elements.Row(styles.Styles{Width: -1, Gap: 16, Align: properties.AlignCenter, Wrap: properties.WrapNoWrap, Padding: properties.SymmetricInset(0, 10)},
		label, control)
}

// rangeValue maps slider position v to d's range, snapped to its step.
func rangeValue(d *settings.Def, v float32) float64 {
	f := d.Min + math.Round(float64(v)*(d.Max-d.Min)/d.Step)*d.Step
	return math.Min(d.Max, math.Max(d.Min, f))
}

// formatRange writes a Range value with as many decimals as its step.
func formatRange(d *settings.Def, f float64) string {
	decimals := 0
	for s := d.Step; s < 0.999 && decimals < 3; s *= 10 {
		decimals++
	}
	return strconv.FormatFloat(f, 'f', decimals, 64)
}
