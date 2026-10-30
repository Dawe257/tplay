package playlist

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func musicTree(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func relativePaths(t *testing.T, p *Playlist, dir string) []string {
	t.Helper()
	var paths []string
	for _, track := range p.Tracks() {
		if !filepath.IsAbs(track.Path) {
			t.Fatalf("track path is not absolute: %q", track.Path)
		}
		path, err := filepath.Rel(dir, track.Path)
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, filepath.ToSlash(path))
	}
	return paths
}

func TestOpenRecursiveSortsByPath(t *testing.T) {
	dir := musicTree(t, "Disc 10/1.flac", "Disc 2/10.WAV", "Disc 2/2.mp3",
		"Disc 2/Bonus/1.opus", "1.ogg", "notes.txt", "Disc 2/cover.jpg")
	p, err := OpenWithOptions(dir, Options{Recursive: true})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"1.ogg", "Disc 2/2.mp3", "Disc 2/10.WAV", "Disc 2/Bonus/1.opus", "Disc 10/1.flac"}
	if got := relativePaths(t, p, dir); !slices.Equal(got, want) {
		t.Fatalf("tracks %q, want %q", got, want)
	}
	if p.Index() != 0 {
		t.Fatalf("start %d, want 0", p.Index())
	}
}

func TestOpenDirectoryDoesNotRecurseByDefault(t *testing.T) {
	dir := musicTree(t, "root.ogg", "nested/song.mp3")
	p, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := relativePaths(t, p, dir); !slices.Equal(got, []string{"root.ogg"}) {
		t.Fatalf("tracks %q, want only root.ogg", got)
	}
}

func TestOpenRecursiveFindsTracksWithEmptyRoot(t *testing.T) {
	dir := musicTree(t, "nested/song.mp3")
	p, err := OpenWithOptions(dir, Options{Recursive: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := relativePaths(t, p, dir); !slices.Equal(got, []string{"nested/song.mp3"}) {
		t.Fatalf("tracks %q, want nested/song.mp3", got)
	}
}

func TestOpenRecursiveEmpty(t *testing.T) {
	dir := musicTree(t, "nested/notes.txt")
	if _, err := OpenWithOptions(dir, Options{Recursive: true}); !errors.Is(err, ErrEmpty) {
		t.Fatalf("got %v, want ErrEmpty", err)
	}
}

func TestOpenRecursiveSymlinkDirectoryTarget(t *testing.T) {
	dir := musicTree(t, "nested/song.mp3")
	link := filepath.Join(t.TempDir(), "music")
	if err := os.Symlink(dir, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	p, err := OpenWithOptions(link, Options{Recursive: true})
	if err != nil {
		t.Fatal(err)
	}
	if p.Len() != 1 || p.Current().FileName() != "song.mp3" {
		t.Fatalf("tracks %v, want song.mp3", p.Tracks())
	}
}

func TestOpenRecursiveMissingTarget(t *testing.T) {
	_, err := OpenWithOptions(filepath.Join(t.TempDir(), "missing"), Options{Recursive: true})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("got %v, want not-exist error", err)
	}
}

func TestOpenRecursiveFileKeepsSiblingPlaylist(t *testing.T) {
	dir := musicTree(t, "1.ogg", "2.ogg", "voice.amr", "nested/3.ogg")
	for _, name := range []string{"2.ogg", "voice.amr"} {
		t.Run(name, func(t *testing.T) {
			p, err := OpenWithOptions(filepath.Join(dir, name), Options{Recursive: true})
			if err != nil {
				t.Fatal(err)
			}
			want := []string{"1.ogg", "2.ogg"}
			if name == "voice.amr" {
				want = []string{"voice.amr", "1.ogg", "2.ogg"}
			}
			if got := relativePaths(t, p, dir); !slices.Equal(got, want) {
				t.Fatalf("tracks %q, want %q", got, want)
			}
			if p.Current().FileName() != name {
				t.Fatalf("current %q, want %q", p.Current().FileName(), name)
			}
		})
	}
}

func TestOpenRecursiveSkipsSymlinks(t *testing.T) {
	dir := musicTree(t, "nested/song.mp3")
	for name, target := range map[string]string{
		"loop.mp3":   dir,
		"alias.mp3":  filepath.Join(dir, "nested", "song.mp3"),
		"broken.mp3": filepath.Join(dir, "missing"),
	} {
		if err := os.Symlink(target, filepath.Join(dir, name)); err != nil {
			t.Skipf("symlinks unavailable: %v", err)
		}
	}
	p, err := OpenWithOptions(dir, Options{Recursive: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := relativePaths(t, p, dir); !slices.Equal(got, []string{"nested/song.mp3"}) {
		t.Fatalf("tracks %q, want nested/song.mp3", got)
	}
}

func TestOpenRecursiveUnreadableSubdirectory(t *testing.T) {
	dir := musicTree(t, "root.ogg", "blocked/song.mp3")
	blocked := filepath.Join(dir, "blocked")
	if err := os.Chmod(blocked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(blocked, 0o755) })
	if _, err := os.ReadDir(blocked); err == nil {
		t.Skip("directory permissions are not enforced for this user")
	}
	if _, err := OpenWithOptions(dir, Options{Recursive: true}); !errors.Is(err, os.ErrPermission) {
		t.Fatalf("got %v, want permission error", err)
	}
}
