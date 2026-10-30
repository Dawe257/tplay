// tplay is a minimal terminal audio player: it plays the given file and lets
// you switch between the audio files in the same directory.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
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
	opts, err := parseArgs(os.Args[1:])
	if errors.Is(err, flag.ErrHelp) {
		fmt.Println(usage)
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "tplay:", err)
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	if opts.version {
		fmt.Println("tplay", buildVersion())
		return
	}
	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "tplay:", err)
		os.Exit(1)
	}
}

const usage = "usage: tplay [-r] [audio file or directory]\n       tplay --version\n\n  -r  include audio files from subdirectories (directory targets only)"

type options struct {
	target    string
	recursive bool
	version   bool
}

func parseArgs(args []string) (options, error) {
	opts := options{target: "."}
	flags := flag.NewFlagSet("tplay", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.BoolVar(&opts.recursive, "r", false, "include audio files from subdirectories")
	flags.BoolVar(&opts.version, "version", false, "print version")
	if err := flags.Parse(args); err != nil {
		return opts, err
	}
	if flags.NArg() > 1 {
		return opts, errors.New("expected at most one audio file or directory")
	}
	if flags.NArg() == 1 {
		opts.target = flags.Arg(0)
	}
	return opts, nil
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

func run(opts options) error {
	pl, err := playlist.OpenWithOptions(opts.target, playlist.Options{Recursive: opts.recursive})
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
