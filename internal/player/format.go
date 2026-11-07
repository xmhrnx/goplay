package player

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

func ftoa(f float64) string { return strconv.FormatFloat(f, 'f', 3, 64) }

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func infOrNo(b bool) string {
	if b {
		return "inf"
	}
	return "no"
}

func idOrNo(id int) string {
	if id <= 0 {
		return "no"
	}
	return strconv.Itoa(id)
}

// picturesDir is where screenshots go: the user's Pictures folder, or
// the home folder.
func picturesDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	if dir := filepath.Join(home, "Pictures"); isDir(dir) {
		return dir
	}
	return home
}

func isDir(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

// ErrNotLoaded is returned by commands that need a file.
var ErrNotLoaded = errors.New("nothing is playing")

// FormatTime formats seconds as m:ss or h:mm:ss.
func FormatTime(secs float64) string {
	if math.IsNaN(secs) || secs < 0 {
		secs = 0
	}
	d := time.Duration(secs * float64(time.Second))
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}
