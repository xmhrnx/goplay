package player

import (
	"fmt"
	"strconv"
	"strings"
)

// MediaInfo collects technical details of what's playing, for the
// statistics panel: name and value pairs in display order.
func (p *Player) MediaInfo() [][2]string {
	get := func(name string) string {
		v, _ := p.h.String(name)
		return v
	}
	var out [][2]string
	add := func(label, value string) {
		if value != "" && value != "no" {
			out = append(out, [2]string{label, value})
		}
	}
	add("File", get("filename"))
	add("Format", get("file-format"))
	if size, ok := p.h.Int("file-size"); ok {
		add("Size", humanBytes(size))
	}
	add("Video codec", get("video-codec"))
	if w, ok := p.h.Int("video-params/w"); ok {
		h, _ := p.h.Int("video-params/h")
		add("Resolution", fmt.Sprintf("%d × %d", w, h))
	}
	add("Pixel format", get("video-params/pixelformat"))
	add("Color", strings.Trim(get("video-params/colormatrix")+" "+get("video-params/primaries")+" "+get("video-params/gamma"), " "))
	if fps, ok := p.h.Double("container-fps"); ok {
		add("Frame rate", fmt.Sprintf("%.3f fps", fps))
	}
	if fps, ok := p.h.Double("estimated-display-fps"); ok {
		add("Display", fmt.Sprintf("%.2f Hz", fps))
	}
	if br, ok := p.h.Int("video-bitrate"); ok && br > 0 {
		add("Video bitrate", fmt.Sprintf("%d kbps", br/1000))
	}
	add("Hardware decoding", get("hwdec-current"))
	add("Audio codec", get("audio-codec-name"))
	if ch, rate := get("audio-params/hr-channels"), get("audio-params/samplerate"); ch != "" && rate != "" {
		add("Audio", ch+" "+rate+" Hz")
	}
	if br, ok := p.h.Int("audio-bitrate"); ok && br > 0 {
		add("Audio bitrate", fmt.Sprintf("%d kbps", br/1000))
	}
	add("Audio output", get("current-ao"))
	if d, ok := p.h.Int("frame-drop-count"); ok {
		add("Dropped frames", strconv.FormatInt(d, 10))
	}
	if d, ok := p.h.Int("decoder-frame-drop-count"); ok {
		add("Decoder drops", strconv.FormatInt(d, 10))
	}
	if c, ok := p.h.Double("demuxer-cache-duration"); ok {
		add("Buffered", fmt.Sprintf("%.1f s", c))
	}
	add("Video filters", get("vf"))
	return out
}
