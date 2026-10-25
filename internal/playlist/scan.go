package playlist

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var audioExts = map[string]bool{
	".aac": true, ".aif": true, ".aiff": true, ".ape": true, ".flac": true,
	".m4a": true, ".mka": true, ".mp3": true, ".oga": true, ".ogg": true,
	".opus": true, ".wav": true, ".wma": true, ".wv": true,
}

// IsAudio reports whether the file name has a known audio extension.
func IsAudio(name string) bool {
	return audioExts[strings.ToLower(filepath.Ext(name))]
}

// Open builds a playlist from the audio files next to target, sorted by name,
// positioned at target. A directory target starts from its first file.
func Open(target string) (*Playlist, error) {
	target, err := filepath.Abs(target)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		tracks, err := scanDir(target)
		if err != nil {
			return nil, err
		}
		return New(tracks, 0)
	}

	tracks, err := scanDir(filepath.Dir(target))
	if err != nil {
		return nil, err
	}
	start := slices.IndexFunc(tracks, func(t Track) bool { return t.Path == target })
	if start < 0 {
		// The opened file always plays, even with an extension not in audioExts.
		tracks, start = append([]Track{{Path: target}}, tracks...), 0
	}
	return New(tracks, start)
}

func scanDir(dir string) ([]Track, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var tracks []Track
	for _, e := range entries {
		if !e.IsDir() && IsAudio(e.Name()) {
			tracks = append(tracks, Track{Path: filepath.Join(dir, e.Name())})
		}
	}
	slices.SortFunc(tracks, func(a, b Track) int {
		return NaturalCompare(a.FileName(), b.FileName())
	})
	return tracks, nil
}
