package ffmpeg

import (
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Prober struct {
	// Bin is the ffprobe executable, "ffprobe" from PATH when empty.
	Bin string
}

func (p Prober) Duration(path string) (time.Duration, error) {
	bin := p.Bin
	if bin == "" {
		bin = "ffprobe"
	}
	out, err := exec.Command(bin, "-v", "error", "-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1", path).Output()
	if err != nil {
		return 0, err
	}
	sec, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(sec * float64(time.Second)), nil
}
