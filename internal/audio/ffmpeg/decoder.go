// Package ffmpeg implements the audio Decoder and Prober ports with the ffmpeg
// and ffprobe command line tools, so every format they know is playable.
package ffmpeg

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"tplay/internal/audio"
)

type Decoder struct {
	// Bin is the ffmpeg executable, "ffmpeg" from PATH when empty.
	Bin string
}

func (d Decoder) Decode(path string, f audio.Format) (io.ReadCloser, error) {
	bin := d.Bin
	if bin == "" {
		bin = "ffmpeg"
	}
	cmd := exec.Command(bin, "-nostdin", "-loglevel", "error", "-i", path, "-vn",
		"-f", "s16le", "-ac", strconv.Itoa(f.Channels), "-ar", strconv.Itoa(f.SampleRate), "-")
	s := &stream{cmd: cmd}
	cmd.Stderr = &s.stderr
	var err error
	if s.stdout, err = cmd.StdoutPipe(); err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return s, nil
}

// stream is the PCM output of a running ffmpeg process.
type stream struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr bytes.Buffer

	waitOnce sync.Once
	waitErr  error
}

// Read turns the end of a failed decoding into ffmpeg's error message.
func (s *stream) Read(p []byte) (int, error) {
	n, err := s.stdout.Read(p)
	if err == io.EOF {
		if werr := s.wait(); werr != nil {
			return n, werr
		}
	}
	return n, err
}

func (s *stream) Close() error {
	s.cmd.Process.Kill()
	s.wait()
	return nil
}

func (s *stream) wait() error {
	s.waitOnce.Do(func() {
		if err := s.cmd.Wait(); err != nil {
			if msg := lastMessage(s.stderr.String()); msg != "" {
				err = errors.New(msg)
			}
			s.waitErr = err
		}
	})
	return s.waitErr
}

// lastMessage returns the last ffmpeg log line without its "[component @ 0x...]" prefix.
func lastMessage(log string) string {
	lines := strings.Split(strings.TrimSpace(log), "\n")
	msg := lines[len(lines)-1]
	if strings.HasPrefix(msg, "[") {
		if i := strings.Index(msg, "] "); i >= 0 {
			msg = msg[i+2:]
		}
	}
	return strings.TrimSpace(msg)
}
