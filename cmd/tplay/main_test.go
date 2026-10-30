package main

import (
	"errors"
	"flag"
	"testing"
)

func TestParseArgs(t *testing.T) {
	for _, tt := range []struct {
		name      string
		args      []string
		target    string
		recursive bool
		version   bool
	}{
		{name: "current directory", target: "."},
		{name: "directory", args: []string{"music"}, target: "music"},
		{name: "recursive current directory", args: []string{"-r"}, target: ".", recursive: true},
		{name: "recursive directory", args: []string{"-r", "my music"}, target: "my music", recursive: true},
		{name: "file", args: []string{"song.ogg"}, target: "song.ogg"},
		{name: "leading dash path", args: []string{"-r", "--", "-music"}, target: "-music", recursive: true},
		{name: "version", args: []string{"--version"}, target: ".", version: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseArgs(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if got.target != tt.target || got.recursive != tt.recursive || got.version != tt.version {
				t.Fatalf("got %+v, want target=%q recursive=%v version=%v", got, tt.target, tt.recursive, tt.version)
			}
		})
	}
}

func TestParseArgsRejectsInvalidInput(t *testing.T) {
	for _, args := range [][]string{{"one", "two"}, {"-r", "one", "two"}, {"--unknown"}} {
		if _, err := parseArgs(args); err == nil {
			t.Errorf("accepted %q", args)
		}
	}
}

func TestParseArgsHelp(t *testing.T) {
	if _, err := parseArgs([]string{"--help"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("got %v, want flag.ErrHelp", err)
	}
}
