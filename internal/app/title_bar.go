package app

import (
	"github.com/mehran9675/giode/elements"
	"github.com/mehran9675/giode/styles"
	"github.com/mehran9675/giode/styles/properties"
)

// titleBar puts the title bar on a solid strip beside the video, or on a
// soft fade from the top edge over it.
func (u *ui) titleBar(overlay bool) elements.Element {
	st := styles.Styles{Width: -1, Height: -1, Background: colSurface}
	if overlay {
		st.Background = ""
		st.BackgroundGradient = properties.Linear(180, "#000000b3", "#00000000")
	}
	return elements.Box(st, u.title)
}

func (u *ui) windowTitle() string {
	st := &u.a.player.State
	switch {
	case u.a.casting():
		return tr("{title} — casting to {device}", "title", st.Title, "device", u.a.cast.device)
	case !st.Idle && st.Title != "":
		return st.Title
	}
	return "GoPlay"
}

// titleButtons are the title bar's extra controls: the app menu.
func (u *ui) titleButtons() []elements.Element {
	more := ghostButton("menu", 18, tr("Menu"), func() { u.openAppMenu() })
	u.appMenuAnchor = &anchor{el: more}
	return []elements.Element{u.appMenuAnchor}
}
