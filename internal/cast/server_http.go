package cast

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func (s *Server) serveMedia(w http.ResponseWriter, r *http.Request) {
	sh, ok := s.lookup(r, "/media/")
	if !ok {
		http.NotFound(w, r)
		return
	}
	// DLNA renderers want these to stream instead of downloading.
	w.Header().Set("transferMode.dlna.org", "Streaming")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if sh.convert {
		w.Header().Set("contentFeatures.dlna.org", "DLNA.ORG_OP=00;DLNA.ORG_CI=1;DLNA.ORG_FLAGS=01700000000000000000000000000000")
		s.streamConverted(w, r, sh)
		return
	}
	w.Header().Set("contentFeatures.dlna.org", "DLNA.ORG_OP=01;DLNA.ORG_CI=0;DLNA.ORG_FLAGS=01700000000000000000000000000000")
	f, err := os.Open(sh.path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", ContentType(sh.path))
	http.ServeContent(w, r, "", fi.ModTime(), f)
}

// streamConverted remuxes the file to fragmented MP4 with FFmpeg,
// copying streams the device plays and transcoding the rest. A "t"
// query parameter starts at that many seconds (devices can't seek in a
// live remux, so seeking restarts the stream there).
func (s *Server) streamConverted(w http.ResponseWriter, r *http.Request, sh *share) {
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		http.Error(w, "FFmpeg is not installed", http.StatusServiceUnavailable)
		return
	}
	args := []string{"-hide_banner", "-loglevel", "error"}
	if t, err := strconv.ParseFloat(r.URL.Query().Get("t"), 64); err == nil && t > 0 {
		args = append(args, "-ss", strconv.FormatFloat(t, 'f', 3, 64))
	}
	args = append(args, "-i", sh.path, "-map", "0:v:0?", "-map", "0:a:0?", "-sn", "-dn")
	if sh.videoCopy {
		args = append(args, "-c:v", "copy")
	} else {
		args = append(args, "-c:v", "libx264", "-preset", "veryfast", "-crf", "21", "-pix_fmt", "yuv420p")
	}
	if sh.audioCopy {
		args = append(args, "-c:a", "copy")
	} else {
		args = append(args, "-c:a", "aac", "-b:a", "192k", "-ac", "2")
	}
	args = append(args, "-movflags", "frag_keyframe+empty_moov+default_base_moof", "-f", "mp4", "pipe:1")
	w.Header().Set("Content-Type", "video/mp4")
	if r.Method == http.MethodHead {
		return
	}
	cmd := exec.CommandContext(r.Context(), ffmpeg, args...)
	cmd.Stdout = w
	cmd.Run()
}

// serveSubs serves subtitles as WebVTT, which both Chromecast and most
// DLNA renderers take.
func (s *Server) serveSubs(w http.ResponseWriter, r *http.Request) {
	sh, ok := s.lookup(r, "/subs/")
	if !ok || sh.subtitles == "" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ext := strings.ToLower(filepath.Ext(sh.subtitles))
	switch {
	case ext == ".vtt":
		http.ServeFile(w, r, sh.subtitles)
	case ext == ".srt":
		f, err := os.Open(sh.subtitles)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		SRTToVTT(w, f)
	default:
		// Anything else (ASS, embedded tracks): FFmpeg converts.
		ffmpeg, err := exec.LookPath("ffmpeg")
		if err != nil {
			http.Error(w, "FFmpeg is not installed", http.StatusServiceUnavailable)
			return
		}
		in, mapArg := sh.subtitles, "0:s:0"
		if sh.subtitles == "embedded" {
			in, mapArg = sh.path, "0:s:"+strconv.Itoa(sh.subIndex)
		}
		cmd := exec.CommandContext(r.Context(), ffmpeg, "-hide_banner", "-loglevel", "error",
			"-i", in, "-map", mapArg, "-f", "webvtt", "pipe:1")
		cmd.Stdout = w
		cmd.Run()
	}
}
