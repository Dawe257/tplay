package playlist

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func tracks(names ...string) []Track {
	var ts []Track
	for _, n := range names {
		ts = append(ts, Track{Path: "/music/" + n})
	}
	return ts
}

func TestNewRejectsEmpty(t *testing.T) {
	if _, err := New(nil, 0); err != ErrEmpty {
		t.Fatalf("got %v, want ErrEmpty", err)
	}
}

func TestNavigationStopsAtEdges(t *testing.T) {
	p, _ := New(tracks("a", "b"), 0)
	if p.Prev() {
		t.Fatal("Prev at the start moved")
	}
	if !p.Next() || p.Index() != 1 {
		t.Fatalf("Next: index %d, want 1", p.Index())
	}
	if p.Next() {
		t.Fatal("Next at the end moved")
	}
}

func TestSkipFollowsLastDirection(t *testing.T) {
	p, _ := New(tracks("a", "b", "c", "d"), 3)
	p.Prev() // at c, moving backward
	if !p.Skip() || p.Index() != 1 {
		t.Fatalf("Skip after Prev: index %d, want 1", p.Index())
	}
	p.Select(2)
	if !p.Skip() || p.Index() != 3 {
		t.Fatalf("Skip after Select: index %d, want 3", p.Index())
	}
}

func TestSelectOutOfRange(t *testing.T) {
	p, _ := New(tracks("a"), 0)
	if p.Select(1) || p.Select(-1) {
		t.Fatal("Select accepted an index out of range")
	}
}

func TestNaturalCompare(t *testing.T) {
	names := []string{"10 c.ogg", "B.wav", "2 b.ogg", "a.mp3", "01 a.ogg"}
	slices.SortFunc(names, NaturalCompare)
	want := []string{"01 a.ogg", "2 b.ogg", "10 c.ogg", "a.mp3", "B.wav"}
	if !slices.Equal(names, want) {
		t.Fatalf("got %q, want %q", names, want)
	}
}

func TestTrackNames(t *testing.T) {
	tr := Track{Path: "/music/01 Song.flac"}
	if tr.Title() != "01 Song" || tr.FileName() != "01 Song.flac" || tr.Dir() != "/music" {
		t.Fatalf("unexpected names: %q %q %q", tr.Title(), tr.FileName(), tr.Dir())
	}
}

func writeFiles(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func fileNames(p *Playlist) []string {
	var names []string
	for _, tr := range p.Tracks() {
		names = append(names, tr.FileName())
	}
	return names
}

func TestOpenFile(t *testing.T) {
	dir := writeFiles(t, "10 c.ogg", "2 b.WAV", "notes.txt", "1 a.opus")
	p, err := Open(filepath.Join(dir, "2 b.WAV"))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"1 a.opus", "2 b.WAV", "10 c.ogg"}; !slices.Equal(fileNames(p), want) {
		t.Fatalf("tracks %q, want %q", fileNames(p), want)
	}
	if p.Index() != 1 {
		t.Fatalf("start %d, want 1", p.Index())
	}
}

func TestOpenUnknownExtensionStillPlays(t *testing.T) {
	dir := writeFiles(t, "a.ogg", "voice.amr")
	p, err := Open(filepath.Join(dir, "voice.amr"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Current().FileName() != "voice.amr" || p.Len() != 2 {
		t.Fatalf("current %q of %d tracks", p.Current().FileName(), p.Len())
	}
}

func TestOpenDirectory(t *testing.T) {
	dir := writeFiles(t, "b.mp3", "a.flac")
	p, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if p.Current().FileName() != "a.flac" {
		t.Fatalf("current %q, want a.flac", p.Current().FileName())
	}
	if _, err := Open(writeFiles(t, "readme.md")); err != ErrEmpty {
		t.Fatalf("empty dir: got %v, want ErrEmpty", err)
	}
}
