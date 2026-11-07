package cast

import (
	"bufio"
	"io"
	"mime"
	"path/filepath"
	"regexp"
	"strings"
)

var srtTime = regexp.MustCompile(`(\d+:\d\d:\d\d),(\d\d\d)`)

// SRTToVTT converts SubRip subtitles to WebVTT.
func SRTToVTT(w io.Writer, r io.Reader) error {
	bw := bufio.NewWriter(w)
	bw.WriteString("WEBVTT\n\n")
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimPrefix(sc.Text(), "\ufeff")
		if strings.Contains(line, "-->") {
			line = srtTime.ReplaceAllString(line, "$1.$2")
		}
		bw.WriteString(line)
		bw.WriteByte('\n')
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return bw.Flush()
}

// ContentType is the MIME type for a media file, by extension.
func ContentType(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".m4v":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mkv":
		return "video/x-matroska"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".ts", ".m2ts", ".mts":
		return "video/mp2t"
	case ".mp3":
		return "audio/mpeg"
	case ".m4a", ".aac":
		return "audio/mp4"
	case ".flac":
		return "audio/flac"
	case ".ogg", ".oga", ".opus":
		return "audio/ogg"
	case ".wav":
		return "audio/wav"
	}
	if t := mime.TypeByExtension(filepath.Ext(path)); t != "" {
		return t
	}
	return "application/octet-stream"
}

// NeedsConversion reports whether a file likely won't play on a cast
// device as is: its container or codecs are outside what Chromecasts
// and most TVs take (MP4/WebM with H.264/VP8/VP9 and AAC/MP3/Opus).
func NeedsConversion(path, videoCodec, audioCodec string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".m4v", ".webm", ".mp3", ".m4a", ".aac":
	default:
		return true
	}
	if videoCodec != "" && !castableVideo(videoCodec) {
		return true
	}
	return audioCodec != "" && !castableAudio(audioCodec)
}

func castableVideo(c string) bool {
	switch strings.ToLower(c) {
	case "h264", "vp8", "vp9":
		return true
	}
	return false
}

func castableAudio(c string) bool {
	switch strings.ToLower(c) {
	case "aac", "mp3", "opus", "vorbis", "flac":
		return true
	}
	return false
}
