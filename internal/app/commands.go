package app

import (
	"fmt"
	"math"
	"strconv"

	"github.com/mehran9675/goplay/internal/actions"
	"github.com/mehran9675/goplay/internal/player"
	"github.com/mehran9675/goplay/internal/settings"
)

// Speed presets offered in menus, and the steps [ and ] move between.
var speedSteps = []float64{0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2, 3, 4}

// registerCommands adds every command the player offers.
func (a *App) registerCommands() {
	p := a.player
	st := &p.State
	hasMedia := func() bool { return !st.Idle }
	hasDuration := func() bool { return !st.Idle && st.Duration > 0 }
	hasPlaylist := func() bool { return len(st.Playlist) > 1 }
	act := func(id, title, cat string, keys []string, run func()) *actions.Action {
		return &actions.Action{ID: id, Title: title, Category: cat, Keys: keys, Run: run}
	}
	when := func(a *actions.Action, enabled func() bool) *actions.Action { a.Enabled = enabled; return a }
	toggle := func(a *actions.Action, checked func() bool) *actions.Action { a.Checked = checked; return a }
	repeat := func(a *actions.Action) *actions.Action { a.Repeat = true; return a }
	k := func(ks ...string) []string { return ks }

	a.actions.Add(
		// File
		act("open", "Open File…", "File", k("Mod+O"), a.openFileDialog),
		act("open_folder", "Open Folder…", "File", k("Mod+Shift+O"), a.openFolderDialog),
		act("open_disc", "Open DVD or Blu-ray…", "File", nil, a.openDisc),
		act("open_url", "Open URL…", "File", k("Mod+U"), func() { a.ui.dialogs.url.show() }),
		act("network_media", "Media on Your Network…", "File", nil, func() { a.ui.dialogs.net.show() }),
		act("paste", "Open Link or File from Clipboard", "File", k("Mod+V"), a.openClipboard),
		act("add_files", "Add Files to Playlist…", "File", nil, a.addFilesDialog),
		when(act("load_subtitle", "Load Subtitle File…", "File", nil, a.loadSubtitleDialog), hasMedia),
		when(act("load_audio", "Load Audio Track…", "File", nil, a.loadAudioDialog), hasMedia),
		when(act("reveal", "Show in Folder", "File", nil, a.revealCurrent), func() bool { return !st.Idle && !isURL(st.Path) }),
		when(act("copy_path", "Copy Path or Link", "File", nil, a.copyCurrentPath), hasMedia),
		when(act("close_media", "Close", "File", k("Mod+W"), p.Stop), hasMedia),
		act("quit", "Quit GoPlay", "File", k("Mod+Q"), a.Quit),

		// Playback
		when(act("play_pause", "Play / Pause", "Playback", k("Space", "K"), a.togglePause), hasMedia),
		when(act("stop", "Stop", "Playback", k("Mod+."), p.Stop), hasMedia),
		when(repeat(act("seek_back", "Back 5 Seconds", "Playback", k("Left", "J"), func() { a.seekBy(-a.seekStep(false)) })), hasDuration),
		when(repeat(act("seek_forward", "Forward 5 Seconds", "Playback", k("Right", "L"), func() { a.seekBy(a.seekStep(false)) })), hasDuration),
		when(repeat(act("seek_back_large", "Back 1 Minute", "Playback", k("Shift+Left"), func() { a.seekBy(-a.seekStep(true)) })), hasDuration),
		when(repeat(act("seek_forward_large", "Forward 1 Minute", "Playback", k("Shift+Right"), func() { a.seekBy(a.seekStep(true)) })), hasDuration),
		when(repeat(act("frame_back", "Previous Frame", "Playback", k(","), func() { p.FrameStep(true) })), hasMedia),
		when(repeat(act("frame_forward", "Next Frame", "Playback", k("."), func() { p.FrameStep(false) })), hasMedia),
		when(act("jump_to", "Go to Time…", "Playback", k("Mod+G"), func() { a.ui.dialogs.jump.show() }), hasDuration),
		when(act("restart", "Start Over", "Playback", k("Home"), func() { a.seekTo(0) }), hasDuration),
		when(act("speed_down", "Slower", "Playback", k("[", "Shift+,"), func() { a.stepSpeed(-1) }), hasMedia),
		when(act("speed_up", "Faster", "Playback", k("]", "Shift+."), func() { a.stepSpeed(1) }), hasMedia),
		when(act("speed_reset", "Normal Speed", "Playback", k("Backspace"), func() { a.setSpeed(1) }), hasMedia),
		when(act("previous", "Previous in Playlist", "Playback", k("PageUp", "Shift+P", "Mod+Left"), a.previous), hasMedia),
		when(act("next", "Next in Playlist", "Playback", k("PageDown", "Shift+N", "Mod+Right"), p.PlaylistNext), hasPlaylist),
		when(act("chapter_previous", "Previous Chapter", "Playback", k("Shift+PageUp"), func() { p.ChapterBy(-1) }), func() bool { return len(st.Chapters) > 0 }),
		when(act("chapter_next", "Next Chapter", "Playback", k("Shift+PageDown"), func() { p.ChapterBy(1) }), func() bool { return len(st.Chapters) > 0 }),
		when(toggle(act("ab_loop", "Repeat Section (A–B)", "Playback", k("B"), a.cycleABLoop), func() bool { return !math.IsNaN(st.ABLoopA) }), hasDuration),
		when(toggle(act("loop_file", "Repeat This File", "Playback", k("R"), func() {
			p.SetLoopFile(!st.LoopFile)
			a.osd(onOff(tr("Repeat"), !st.LoopFile))
		}), func() bool { return st.LoopFile }), hasMedia),
		toggle(act("loop_playlist", "Repeat Playlist", "Playback", nil, func() {
			p.SetLoopPlaylist(!st.LoopPlaylist)
			a.osd(onOff(tr("Repeat playlist"), !st.LoopPlaylist))
		}), func() bool { return st.LoopPlaylist }),
		when(act("shuffle", "Shuffle Playlist", "Playback", nil, func() { p.PlaylistShuffle(); a.osd(tr("Shuffled")) }), hasPlaylist),

		// Audio
		repeat(act("volume_up", "Volume Up", "Audio", k("Up", "=", "+"), func() { a.addVolume(a.settings.Float(settings.KeyVolumeStep)) })),
		repeat(act("volume_down", "Volume Down", "Audio", k("Down", "-"), func() { a.addVolume(-a.settings.Float(settings.KeyVolumeStep)) })),
		toggle(act("mute", "Mute", "Audio", k("M"), a.toggleMute), func() bool { return st.Muted }),
		when(act("audio_next", "Next Audio Track", "Audio", k("A"), func() { a.cycleTrack(player.Audio) }), func() bool { return len(st.TracksOf(player.Audio)) > 1 }),
		when(act("audio_delay_up", "Delay Audio +0.1 s", "Audio", k("Mod+Shift+="), func() { a.addAudioDelay(0.1) }), hasMedia),
		when(act("audio_delay_down", "Delay Audio −0.1 s", "Audio", k("Mod+Shift+-"), func() { a.addAudioDelay(-0.1) }), hasMedia),

		// Subtitles
		when(toggle(act("subs_toggle", "Show Subtitles", "Subtitles", k("C"), a.toggleSubtitles), func() bool {
			_, on := st.SelectedTrack(player.Subtitle)
			return on
		}), func() bool { return len(st.TracksOf(player.Subtitle)) > 0 }),
		when(act("subs_next", "Next Subtitle Track", "Subtitles", k("Shift+C"), func() { a.cycleTrack(player.Subtitle) }), func() bool { return len(st.TracksOf(player.Subtitle)) > 0 }),
		when(repeat(act("sub_delay_up", "Subtitles Later (+0.1 s)", "Subtitles", k("X"), func() { a.addSubDelay(0.1) })), hasMedia),
		when(repeat(act("sub_delay_down", "Subtitles Earlier (−0.1 s)", "Subtitles", k("Z"), func() { a.addSubDelay(-0.1) })), hasMedia),
		when(repeat(act("sub_bigger", "Bigger Subtitles", "Subtitles", k("Alt+Up"), func() { a.scaleSubs(0.1) })), hasMedia),
		when(repeat(act("sub_smaller", "Smaller Subtitles", "Subtitles", k("Alt+Down"), func() { a.scaleSubs(-0.1) })), hasMedia),
		when(act("sub_previous_line", "Previous Subtitle Line", "Subtitles", k("Mod+Shift+Left"), func() { p.SubStep(-1) }), hasMedia),
		when(act("sub_next_line", "Next Subtitle Line", "Subtitles", k("Mod+Shift+Right"), func() { p.SubStep(1) }), hasMedia),

		// Video
		toggle(act("fullscreen", "Fullscreen", "Video", k("F", "F11"), func() { a.setFullscreen(!a.g.Fullscreen()) }), a.g.Fullscreen),
		toggle(act("maximize", "Maximize", "Video", nil, func() { windowCtl{a}.ToggleMaximize() }), a.g.Maximized),
		toggle(act("pip", "Picture in Picture", "Video", k("P"), a.togglePiP), func() bool { return a.ui.pip }),
		toggle(act("always_on_top", "Always on Top", "Video", k("T"), a.toggleOnTop), func() bool { return a.g.AlwaysOnTop() }),
		when(act("screenshot", "Take Screenshot", "Video", k("S"), func() { a.screenshot("subtitles") }), hasMedia),
		when(act("screenshot_clean", "Take Screenshot Without Subtitles", "Video", k("Shift+S"), func() { a.screenshot("video") }), hasMedia),
		when(act("rotate", "Rotate 90°", "Video", k("Mod+R"), a.rotate), hasMedia),
		when(repeat(act("zoom_in", "Zoom In", "Video", k("Mod+="), func() { a.zoomBy(0.1) })), hasMedia),
		when(repeat(act("zoom_out", "Zoom Out", "Video", k("Mod+-"), func() { a.zoomBy(-0.1) })), hasMedia),
		when(act("zoom_reset", "Reset Zoom", "Video", k("Mod+0"), a.resetZoom), hasMedia),
		when(act("fit_cycle", "Change Picture Size", "Video", nil, a.cycleFit), hasMedia),
		when(toggle(act("deinterlace", "Deinterlace", "Video", k("D"), func() { a.toggleSetting("deinterlace", "Deinterlace") }),
			func() bool { return a.settings.Bool("deinterlace") }), hasMedia),
		when(act("next_video", "Next Video Track", "Video", nil, func() { a.cycleTrack(player.Video) }), func() bool { return len(st.TracksOf(player.Video)) > 1 }),
		when(toggle(act("info", "Media Information", "Video", k("I"), func() { a.ui.toggleInfo() }), func() bool { return a.ui.infoOpen }), hasMedia),

		// 3D & VR
		when(act("stereo_off", "3D Off (2D)", "3D & VR", nil, func() { a.setStereoLayout(player.Layout2D) }), hasMedia),
		when(act("stereo_sbs", "Side-by-Side 3D", "3D & VR", nil, func() { a.setStereoLayout(player.LayoutSBSHalf) }), hasMedia),
		when(act("stereo_tb", "Top-and-Bottom 3D", "3D & VR", nil, func() { a.setStereoLayout(player.LayoutTBHalf) }), hasMedia),
		when(act("vr_360", "360° Video", "3D & VR", nil, func() { a.setProjection(player.ProjectionEquirect, player.VRMono) }), hasMedia),
		when(act("vr_180", "VR180 Video", "3D & VR", nil, func() { a.setProjection(player.ProjectionHalf, player.VRSideBySide) }), hasMedia),
		when(act("vr_reset", "Reset 360° View", "3D & VR", k("Mod+Shift+0"), a.resetVRView), a.vrActive),

		// Cast
		toggle(act("cast", "Cast to TV or Speaker…", "Cast", k("Mod+Shift+C"), func() { a.ui.dialogs.cast.show() }), a.casting),
		when(act("cast_stop", "Stop Casting", "Cast", nil, a.stopCasting), a.casting),

		// View
		toggle(act("playlist", "Playlist", "View", k("Mod+L"), func() { a.ui.togglePanel(panelPlaylist) }), func() bool { return a.ui.panelOpen && a.ui.panelTab == panelPlaylist }),
		when(toggle(act("chapters", "Chapters", "View", nil, func() { a.ui.togglePanel(panelChapters) }), func() bool { return a.ui.panelOpen && a.ui.panelTab == panelChapters }),
			func() bool { return len(st.Chapters) > 0 }),
		act("palette", "Command Palette…", "View", k("Mod+K", "Mod+Shift+P"), func() { a.ui.dialogs.palette.show() }),
		act("settings", "Settings…", "View", k("Mod+,"), func() { a.ui.dialogs.settings.show("") }),
		act("shortcuts", "Keyboard Shortcuts", "View", k("Shift+/", "F1"), func() { a.ui.dialogs.settings.show("Shortcuts") }),
		act("log", "mpv Log", "View", nil, func() { a.ui.dialogs.log.show() }),
		act("config_folder", "Open Configuration Folder", "View", nil, a.openConfigFolder),
		act("about", "About GoPlay", "View", nil, func() { a.ui.dialogs.about.show() }),
	)

	// 1 to 9 jump to 10% to 90%, 0 to the start.
	for i := 0; i <= 9; i++ {
		frac := float64(i) / 10
		a.actions.Add(&actions.Action{
			ID: "seek_percent_" + strconv.Itoa(i), Title: fmt.Sprintf("Jump to %d%%", i*10), Category: "Playback",
			Keys: []string{strconv.Itoa(i)}, Hidden: true, Enabled: hasDuration,
			Run: func() { a.seekTo(frac * st.Duration) },
		})
	}
}

// applyKeyBindings loads the user's custom shortcuts.
func (a *App) applyKeyBindings() {
	for _, act := range a.actions.All() {
		if keys, ok := a.settings.Keys(act.ID); ok {
			a.actions.SetCustom(act.ID, keys)
		}
	}
}

func onOff(what string, on bool) string {
	if on {
		return tr("{what}: on", "what", what)
	}
	return tr("{what}: off", "what", what)
}
