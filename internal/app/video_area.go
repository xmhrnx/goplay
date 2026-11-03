package app

import (
	"path/filepath"
	"time"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/spinner"
	"github.com/xmhrnx/giode/elements"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/player"
	"github.com/mehran9675/goplay/internal/settings"
)

// videoArea is the picture's rectangle: it takes clicks (pause,
// fullscreen, the context menu), the wheel and 360° dragging, and shows
// what goes over the video.
type videoArea struct {
	u     *ui
	press *elements.Pressable

	wheelTag struct{}
	vrDrag   gesture.Drag
	vrLast   f32.Point

	welcome *welcomeScreen
}

func newVideoArea(u *ui) *videoArea {
	v := &videoArea{u: u, welcome: newWelcomeScreen(u)}
	a := u.a
	v.press = elements.NewPressable(v.content).Label(tr("Video")).Focusable(false).
		Styles(styles.Styles{Width: -1, Height: -1, Cursor: properties.CursorDefault}).
		OnClick(func(e elements.ClickEvent) {
			st := &a.player.State
			// OnClick sees every button; the right one opens the menu.
			if e.Button != elements.MousePrimary || st.Idle || a.vrActive() || u.modalOpen() {
				return
			}
			switch e.Clicks {
			case 1:
				if a.settings.Bool(settings.KeyClickToPause) {
					a.togglePause()
				}
			case 2:
				// The first click of the pair paused; undo it.
				if a.settings.Bool(settings.KeyClickToPause) {
					a.togglePause()
					u.flash = ""
				}
				a.setFullscreen(!a.g.Fullscreen())
			}
		}).
		OnSecondaryClick(func(e elements.ClickEvent) {
			u.openMenuAt(u.contextMenu(), u.videoRect.Min.Add(e.Position))
		})
	return v
}

func (v *videoArea) Layout(gtx layout.Context) layout.Dimensions {
	size := gtx.Constraints.Max
	dims := v.press.Layout(gtx)
	// The wheel and 360° dragging sit above the click area and pass
	// events through to it.
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	event.Op(gtx.Ops, &v.wheelTag)
	v.handleWheel(gtx)
	if v.u.a.vrActive() {
		v.vrDrag.Add(gtx.Ops)
		pointer.CursorGrab.Add(gtx.Ops)
		v.handleVRDrag(gtx)
	}
	pass.Pop()
	area.Pop()
	return dims
}

func (v *videoArea) handleWheel(gtx layout.Context) {
	a := v.u.a
	for {
		ev, ok := gtx.Event(pointer.Filter{Target: &v.wheelTag, Kinds: pointer.Scroll, ScrollY: pointer.ScrollRange{Min: -1e6, Max: 1e6}})
		if !ok {
			return
		}
		e, ok := ev.(pointer.Event)
		if !ok || e.Scroll.Y == 0 || a.player.State.Idle {
			continue
		}
		steps := e.Scroll.Y / float32(gtx.Dp(48))
		switch {
		case a.vrActive():
			a.vrZoom(steps)
		case a.settings.Get(settings.KeyWheel) == "volume":
			a.addVolume(-float64(steps) * a.settings.Float(settings.KeyVolumeStep))
		case a.settings.Get(settings.KeyWheel) == "seek":
			a.seekBy(-float64(steps) * a.seekStep(false))
		}
	}
}

func (v *videoArea) handleVRDrag(gtx layout.Context) {
	a := v.u.a
	for {
		e, ok := v.vrDrag.Update(gtx.Metric, gtx.Source, gesture.Both)
		if !ok {
			return
		}
		switch e.Kind {
		case pointer.Press:
			v.vrLast = e.Position
			a.stereo.dragging = true
		case pointer.Drag:
			d := e.Position.Sub(v.vrLast)
			v.vrLast = e.Position
			a.vrDrag(d.X, d.Y)
		case pointer.Release, pointer.Cancel:
			a.vrDragEnd()
		}
	}
}

// content is everything drawn over the picture.
func (v *videoArea) content(elements.PressState) elements.Element {
	u, a := v.u, v.u.a
	st := &a.player.State
	var layers []elements.Element

	switch {
	case a.casting():
		layers = append(layers, castScreen(a))
	case st.Idle:
		layers = append(layers, v.welcome.element(false))
	case a.finished():
		layers = append(layers, v.welcome.element(true))
	case !st.HasVideo() && !st.Loading:
		layers = append(layers, audioOnly(st))
	}
	if (st.Loading || st.Buffering) && !st.Idle && !a.casting() {
		layers = append(layers, centered(elements.Box(styles.Styles{
			Background: "#000000a0", BorderRadius: 28, Padding: properties.UniformInset(14),
		}, spinner.Spinner(styles.Styles{Width: 28, Height: 28, Color: colText}))))
	}
	if d, ok := a.g.FileDrag(); ok && d.Kind != giode.FileDragLeave {
		layers = append(layers, dropTarget(!st.Idle))
	}
	if u.flash != "" {
		if age := time.Since(u.flashAt); age < flashDuration {
			f := float32(age) / float32(flashDuration)
			layers = append(layers, centered(elements.Box(styles.Styles{
				Background: "#00000080", BorderRadius: 40, Padding: properties.UniformInset(18),
				Opacity:   properties.Opacity(max(0.01, 1-f)),
				Transform: properties.Transform{ScaleX: 1 + f*0.3, ScaleY: 1 + f*0.3},
			}, ic(u.flash, 40, colHighlight))))
		} else {
			u.flash = ""
		}
	}
	if u.osdText != "" {
		if time.Now().Before(u.osdUntil) {
			layers = append(layers, osdView(u))
		} else {
			u.osdText = ""
		}
	}
	if u.infoOpen && !st.Idle {
		layers = append(layers, infoPanel(u))
	}
	return elements.Stack(styles.Styles{Width: -1, Height: -1}, layers...)
}

// centered places el in the middle of the video area.
func centered(el elements.Element) elements.Element {
	return elements.Box(styles.Styles{Position: properties.PositionAbsolute, Top: 0, Left: 0, Right: 0, Bottom: 0,
		Justify: properties.JustifyCenter, Align: properties.AlignCenter, Width: -1, Height: -1}, el)
}

// displayTitle is the playing file's title, or its name.
func displayTitle(st *player.State) string {
	if st.Title != "" {
		return st.Title
	}
	return filepath.Base(st.Path)
}
