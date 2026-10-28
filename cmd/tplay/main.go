// tplay is a minimal terminal audio player: it plays the given file and lets
// you switch between the audio files in the same directory.
package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dzhenetl/tplay/internal/audio"
	"github.com/dzhenetl/tplay/internal/audio/ffmpeg"
	"github.com/dzhenetl/tplay/internal/audio/otoout"
	"github.com/dzhenetl/tplay/internal/playlist"
	"github.com/dzhenetl/tplay/internal/ui"
)

// version is set at build time with -ldflags "-X main.version=...". Without
// it (go install), the module version from the build info is used.
var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println("tplay", buildVersion())
		return
	}
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: tplay <audio file or directory>\n       tplay --version")
		os.Exit(2)
	}
	if err := run(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "tplay:", err)
		os.Exit(1)
	}
}

// buildVersion returns the version without the leading "v", as the Makefile does.
func buildVersion() string {
	if version != "dev" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return version
	}
	return strings.TrimPrefix(info.Main.Version, "v")
}

func run(target string) error {
	pl, err := playlist.Open(target)
	if err != nil {
		return err
	}
	output, err := otoout.New(audio.DefaultFormat)
	if err != nil {
		return err
	}
	player := audio.NewPlayer(ffmpeg.Decoder{}, ffmpeg.Prober{}, output, audio.DefaultFormat)
	_, err = tea.NewProgram(ui.New(player, pl), tea.WithAltScreen()).Run()
	return err
}
