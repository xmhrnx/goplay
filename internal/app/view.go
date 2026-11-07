package app

import (
	"image"
	"runtime"
	"time"

	"gioui.org/f32"
	"gioui.org/gesture"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/drawer"
	"github.com/xmhrnx/giode/components/menu"
	"github.com/xmhrnx/giode/components/titlebar"
	"github.com/xmhrnx/giode/components/toaster"
	"github.com/xmhrnx/giode/styles"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/goplay/internal/settings"
)

// ui is the view state: components created once and laid out every
// frame, and the transient state of the chrome (what's hovered, open,
// fading).
type ui struct {
	a *App

	fb        image.Point
	scale     float32
	videoRect image.Rectangle
	// marginTop and marginBottom are the video margins last sent to mpv.
	marginTop, marginBottom float64

	title *titlebar.TitleBar
	bar   *controlBar

	// The chrome in overlay mode (maximized, fullscreen, picture in
	// picture) shows while the pointer is over its strip.
	topHover, bottomHover gesture.Hover
	topShown, bottomShown bool
	lastMove              time.Time
	lastPointer           f32.Point
	cursorHidden          bool
	pointerTag            struct{}

	pip         bool
	pipRestore  giode.Rect // the window's place and size before, in dp
	pipHadFrame bool       // whether pipRestore's position is known
	pipWasOnTop bool

	video *videoArea

	menu *menu.Menu // the open context or popover menu

	panelOpen bool
	panelTab  int
	panel     *sidePanel
	drawer    *drawer.Drawer

	osdText  string
	osdLevel float64 // 0 to 1 for a level bar, -1 for none
	osdUntil time.Time
	flash    string
	flashAt  time.Time

	infoOpen bool
	infoRows [][2]string
	infoNext time.Time

	dialogs *dialogs

	toasts *toaster.Toaster

	appMenuAnchor *anchor

	// flipH and flipV mirror the picture (per file).
	flipH, flipV bool
}

func newUI(a *App) *ui {
	u := &ui{a: a, osdLevel: -1, lastMove: time.Now()}
	// The bar itself is transparent: titleBar puts it on a solid strip
	// or a fade, depending on the mode.
	u.title = titlebar.New(windowCtl{a}, "GoPlay", styles.Styles{
		Background: "transparent",
		Color:      colText,
		FontSize:   13,
		Height:     properties.Height(u.titleHeightDp()),
	})
	if runtime.GOOS == "darwin" {
		u.title.Variant(titlebar.VariantMac)
	}
	u.title.Trailing(u.titleButtons()...)
	u.bar = newControlBar(u)
	u.video = newVideoArea(u)
	u.panel = newSidePanel(u)
	u.drawer = drawer.New(&u.panelOpen, drawer.Right, panelWidth, u.panel.content, styles.Styles{
		Background: colSurface,
	}).Animated(false)
	u.dialogs = newDialogs(u)
	u.toasts = toaster.New(styles.Styles{Background: colSurface2, BorderColor: colBorder, Color: colText, FontSize: 13}).
		Placement(toaster.TopCenter).Max(3)
	return u
}

// rebuildUI makes the interface again, as for a new language: its parts
// are made once with their text. Where the window and picture stand,
// and what's open, carry over.
func (a *App) rebuildUI() {
	old := a.ui
	u := newUI(a)
	u.fb, u.scale, u.videoRect = old.fb, old.scale, old.videoRect
	u.marginTop, u.marginBottom = old.marginTop, old.marginBottom
	u.lastMove, u.lastPointer = old.lastMove, old.lastPointer
	u.pip, u.pipRestore, u.pipHadFrame, u.pipWasOnTop = old.pip, old.pipRestore, old.pipHadFrame, old.pipWasOnTop
	u.panelOpen, u.panelTab = old.panelOpen, old.panelTab
	u.panel.tabs.SetSelected(old.panelTab)
	u.infoOpen = old.infoOpen
	u.flipH, u.flipV = old.flipH, old.flipV
	if old.dialogs.settings.open {
		u.dialogs.settings.show(old.dialogs.settings.page)
	}
	a.ui = u
}

func (u *ui) titleHeightDp() int {
	if runtime.GOOS == "darwin" {
		return titleBarMacH
	}
	return titleBarH
}

func (u *ui) dp(v int) int { return int(float32(v)*u.scale + 0.5) }

func (u *ui) titleHeightPx() int { return u.dp(u.titleHeightDp()) }
func (u *ui) barHeightPx() int   { return u.dp(controlBarH) }

// overlay reports whether the chrome floats over the video (maximized,
// fullscreen, picture in picture) rather than sitting beside it.
func (u *ui) overlay() bool {
	g := u.a.g
	return g.Maximized() || g.Fullscreen() || u.pip
}

// beginFrame records the frame's size and refreshes what depends on
// time.
func (u *ui) beginFrame(fb image.Point, scale float32) {
	u.fb, u.scale = fb, scale
	if u.overlay() {
		u.videoRect = image.Rectangle{Max: fb}
	} else {
		u.videoRect = image.Rect(0, u.titleHeightPx(), fb.X, fb.Y-u.barHeightPx())
	}
	if u.infoOpen && !time.Now().Before(u.infoNext) {
		u.infoRows = u.a.player.MediaInfo()
		u.infoNext = time.Now().Add(time.Second)
	}
	u.a.vrTick()
}

// nextTimer is the next moment the UI changes on its own.
func (u *ui) nextTimer() time.Time {
	var next time.Time
	consider := func(t time.Time) {
		if !t.IsZero() && t.After(time.Now().Add(-time.Second)) && (next.IsZero() || t.Before(next)) {
			next = t
		}
	}
	if u.osdText != "" {
		consider(u.osdUntil)
	}
	if u.flash != "" {
		consider(u.flashAt.Add(flashDuration))
	}
	if u.infoOpen {
		consider(u.infoNext)
	}
	if u.overlay() && !u.cursorHidden {
		consider(u.lastMove.Add(u.hideDelay()))
	}
	consider(u.a.vrTimer())
	return next
}

func (u *ui) hideDelay() time.Duration {
	s := u.a.settings.Float(settings.KeyHideDelay)
	if s <= 0 {
		s = 2
	}
	return time.Duration(s * float64(time.Second))
}

// modalOpen reports whether a dialog takes the keyboard.
func (u *ui) modalOpen() bool { return u.dialogs.anyOpen() }

// closeTopmost closes the innermost open thing; false if none was.
func (u *ui) closeTopmost() bool {
	switch {
	case u.menu != nil && u.menu.Opened():
		u.menu.Close()
	case u.dialogs.closeTop():
	case u.panelOpen:
		u.panelOpen = false
	case u.infoOpen:
		u.infoOpen = false
	default:
		return false
	}
	u.a.dirty = true
	return true
}

func (u *ui) togglePanel(tab int) {
	if u.panelOpen && u.panelTab == tab {
		u.panelOpen = false
		return
	}
	u.panelTab = tab
	u.panelOpen = true
	u.panel.opened()
}

func (u *ui) toggleInfo() {
	u.infoOpen = !u.infoOpen
	u.infoNext = time.Time{}
}

// fileLoaded resets per-file view state.
func (u *ui) fileLoaded() {
	u.a.Post(u.a.sizeToVideo)
	u.panel.scrollToCurrent()
}
