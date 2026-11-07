package settings

// Keys of the settings the app (not mpv) acts on.
const (
	KeyLanguage        = "language"
	KeyAlwaysOnTop     = "always_on_top"
	KeyResizeToVideo   = "resize_to_video"
	KeyRemainingTime   = "remaining_time"
	KeyHideDelay       = "hide_delay"
	KeyClickToPause    = "click_to_pause"
	KeyWheel           = "wheel"
	KeyOSD             = "osd"
	KeyRememberVolume  = "remember_volume"
	KeyPauseMinimized  = "pause_minimized"
	KeyQueueFolder     = "queue_folder"
	KeySeekStep        = "seek_step"
	KeySeekStepLarge   = "seek_step_large"
	KeyVolumeStep      = "volume_step"
	KeyAudioDevice     = "audio_device"
	KeyAuto3D          = "auto_3d"
	KeyStereoView      = "stereo_view"
	KeyVRFOV           = "vr_fov"
	KeyVRSensitivity   = "vr_sensitivity"
	KeyVRInvert        = "vr_invert"
	KeyCastConvert     = "cast_convert"
	KeyCastSubtitles   = "cast_subtitles"
	KeyMediaServers    = "media_servers"
	KeyPassKeys        = "pass_keys"
	KeyMPVConf         = "mpv_conf"
	KeyLogLevel        = "log_level"
	KeyExtraMPVOptions = "extra_mpv"

	// KeyVolume stores the last volume (percent) for KeyRememberVolume.
	// It has no Def: it isn't shown in the dialog.
	KeyVolume = "volume"
	// KeySubtitleChoice stores the subtitle track last picked, as a
	// player.TrackChoice in JSON, to pick in every file after.
	KeySubtitleChoice = "subtitle_choice"
)
