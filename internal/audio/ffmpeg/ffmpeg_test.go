package ffmpeg

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/dzhenetl/tplay/internal/audio"
)

func TestLastMessage(t *testing.T) {
	log := "first line\n[in#0 @ 0x55d0] Error opening input files: Invalid data\n"
	if got := lastMessage(log); got != "Error opening input files: Invalid data" {
		t.Fatalf("got %q", got)
	}
}

// These tests run the real ffmpeg tools and are skipped where they are missing.
func requireTools(t *testing.T) {
	t.Helper()
	for _, bin := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(bin); err != nil {
			t.Skip(bin + " not installed")
		}
	}
}

func sine(t *testing.T, seconds string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sine.wav")
	out, err := exec.Command("ffmpeg", "-loglevel", "error", "-f", "lavfi",
		"-i", "sine=f=440:d="+seconds, path).CombinedOutput()
	if err != nil {
		t.Fatalf("make test file: %v: %s", err, out)
	}
	return path
}

func TestDecodeAndProbe(t *testing.T) {
	requireTools(t)
	path := sine(t, "1")

	d, err := Prober{}.Duration(path)
	if err != nil || d.Round(time.Millisecond) != time.Second {
		t.Fatalf("duration %v, %v; want 1s", d, err)
	}

	s, err := Decoder{}.Decode(path, audio.DefaultFormat)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	pcm, err := io.ReadAll(s)
	if err != nil {
		t.Fatal(err)
	}
	if got := audio.DefaultFormat.Duration(int64(len(pcm))).Round(time.Millisecond); got != time.Second {
		t.Fatalf("decoded %v of audio, want 1s", got)
	}
}

func TestDecodeBrokenFile(t *testing.T) {
	requireTools(t)
	path := filepath.Join(t.TempDir(), "broken.mp3")
	os.WriteFile(path, []byte("not audio"), 0o644)

	s, err := Decoder{}.Decode(path, audio.DefaultFormat)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := io.ReadAll(s); err == nil {
		t.Fatal("broken file decoded without error")
	}
}

func TestCloseStopsDecoding(t *testing.T) {
	requireTools(t)
	s, err := Decoder{}.Decode(sine(t, "60"), audio.DefaultFormat)
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		io.Copy(io.Discard, s)
		close(done)
	}()
	s.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("reading did not stop after Close")
	}
}
