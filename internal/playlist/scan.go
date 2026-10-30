package playlist

import (
	"io/fs"
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
	return OpenWithOptions(target, Options{})
}

// Options controls how directory targets are scanned.
type Options struct {
	Recursive bool
}

// OpenWithOptions builds a playlist like Open. Recursive directory scans include
// audio files in subdirectories, sorted naturally by path, and skip symlinks
// encountered inside the tree. A symlink used as the target is resolved first.
// Options do not change how individual file targets are opened.
func OpenWithOptions(target string, opts Options) (*Playlist, error) {
	target, err := filepath.Abs(target)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(target)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		scan := scanDir
		if opts.Recursive {
			scan = scanRecursive
		}
		tracks, err := scan(target)
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

func scanRecursive(dir string) ([]Track, error) {
	// WalkDir does not follow even its root when the root is a symlink.
	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, err
	}
	var tracks []Track
	err = filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() && IsAudio(entry.Name()) {
			tracks = append(tracks, Track{Path: path})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.SortFunc(tracks, func(a, b Track) int {
		if cmp := NaturalCompare(a.Path, b.Path); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.Path, b.Path)
	})
	return tracks, nil
}
