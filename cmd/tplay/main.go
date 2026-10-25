// tplay is a minimal terminal audio player: it plays the given file and lets
// you switch between the audio files in the same directory.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"tplay/internal/audio"
	"tplay/internal/audio/ffmpeg"
	"tplay/internal/audio/otoout"
	"tplay/internal/playlist"
	"tplay/internal/ui"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: tplay <audio file or directory>")
		os.Exit(2)
	}
	if err := run(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "tplay:", err)
		os.Exit(1)
	}
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
