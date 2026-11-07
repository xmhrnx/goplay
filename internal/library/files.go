package library

import (
	"path/filepath"
	"strings"
)

// MediaExts are the file extensions goplay treats as media when
// scanning folders.
var MediaExts = map[string]bool{
	".mp4": true, ".m4v": true, ".mkv": true, ".mov": true, ".avi": true, ".webm": true,
	".ts": true, ".m2ts": true, ".mts": true, ".flv": true, ".wmv": true, ".mpg": true,
	".mpeg": true, ".ogv": true, ".3gp": true, ".vob": true, ".divx": true, ".rmvb": true,
	".mp3": true, ".flac": true, ".wav": true, ".ogg": true, ".oga": true, ".opus": true,
	".m4a": true, ".aac": true, ".mka": true, ".wma": true, ".alac": true, ".ape": true,
	".iso": true,
}

// SubtitleExts are subtitle file extensions.
var SubtitleExts = map[string]bool{
	".srt": true, ".ass": true, ".ssa": true, ".vtt": true, ".sub": true, ".idx": true,
	".sup": true, ".smi": true, ".lrc": true,
}

// IsMedia reports whether path has a media file extension.
func IsMedia(path string) bool { return MediaExts[strings.ToLower(filepath.Ext(path))] }

// IsSubtitle reports whether path has a subtitle file extension.
func IsSubtitle(path string) bool { return SubtitleExts[strings.ToLower(filepath.Ext(path))] }

// IsURL reports whether s looks like a URL rather than a path.
func IsURL(s string) bool {
	i := strings.Index(s, "://")
	return i > 1 && !strings.ContainsAny(s[:i], `\/.`)
}
