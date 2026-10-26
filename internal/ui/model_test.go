package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Dawe257/tplay/internal/audio"
	"github.com/Dawe257/tplay/internal/playlist"
)

// fakePlayer records loads; tests set the status it reports.
type fakePlayer struct {
	loaded []string
	status audio.Status
	seeks  []time.Duration
}

func (p *fakePlayer) Load(path string) {
	p.loaded = append(p.loaded, path)
	p.status = audio.Status{Volume: p.status.Volume}
}
func (p *fakePlayer) Stop()                  {}
func (p *fakePlayer) TogglePause()           { p.status.Paused = !p.status.Paused }
func (p *fakePlayer) Seek(to time.Duration)  { p.seeks = append(p.seeks, to) }
func (p *fakePlayer) SeekBy(d time.Duration) { p.Seek(p.status.Position + d) }
func (p *fakePlayer) SetVolume(v float64)    { p.status.Volume = v }
func (p *fakePlayer) Status() audio.Status   { return p.status }

func newTestModel(t *testing.T, start int, names ...string) (*Model, *fakePlayer) {
	t.Helper()
	var tracks []playlist.Track
	for _, n := range names {
		tracks = append(tracks, playlist.Track{Path: "/music/" + n})
	}
	pl, err := playlist.New(tracks, start)
	if err != nil {
		t.Fatal(err)
	}
	p := &fakePlayer{}
	m := New(p, pl)
	m.Init()
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
	return m, p
}

func press(m *Model, keys string) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(keys)}
	switch keys {
	case " ":
		msg = tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")}
	case "enter":
		msg = tea.KeyMsg{Type: tea.KeyEnter}
	case "down":
		msg = tea.KeyMsg{Type: tea.KeyDown}
	}
	m.Update(msg)
}

func finish(m *Model, p *fakePlayer, err error) {
	p.status.Finished, p.status.Err = true, err
	m.Update(tickMsg{})
}

func lastLoaded(p *fakePlayer) string {
	return p.loaded[len(p.loaded)-1]
}

func TestStartsWithOpenedTrack(t *testing.T) {
	_, p := newTestModel(t, 1, "a.ogg", "b.ogg")
	if lastLoaded(p) != "/music/b.ogg" {
		t.Fatalf("loaded %q", p.loaded)
	}
}

func TestFinishedTrackAdvancesAndStopsAtTheEnd(t *testing.T) {
	m, p := newTestModel(t, 0, "a.ogg", "b.ogg")
	finish(m, p, nil)
	if lastLoaded(p) != "/music/b.ogg" {
		t.Fatalf("loaded %q", p.loaded)
	}
	finish(m, p, nil)
	if !m.ended || len(p.loaded) != 2 {
		t.Fatalf("ended=%v loaded %q", m.ended, p.loaded)
	}
	press(m, " ")
	if len(p.loaded) != 3 || m.ended {
		t.Fatal("space at the end did not replay the track")
	}
}

func TestFailedTrackIsSkippedBackwardAfterPrev(t *testing.T) {
	m, p := newTestModel(t, 2, "a.ogg", "broken.mp3", "c.ogg")
	press(m, "p")
	finish(m, p, errors.New("invalid data"))
	if lastLoaded(p) != "/music/a.ogg" {
		t.Fatalf("loaded %q, want a.ogg after skipping back", p.loaded)
	}
	if !strings.Contains(m.err, "broken.mp3: invalid data") {
		t.Fatalf("error %q", m.err)
	}
}

func TestPrevRestartsAfterThreeSeconds(t *testing.T) {
	m, p := newTestModel(t, 1, "a.ogg", "b.ogg")
	p.status.Position = 10 * time.Second
	m.Update(tickMsg{})
	press(m, "p")
	if len(p.loaded) != 1 || len(p.seeks) != 1 || p.seeks[0] != 0 {
		t.Fatalf("loaded %q seeks %v, want a restart", p.loaded, p.seeks)
	}
}

func TestEnterPlaysSelectedTrack(t *testing.T) {
	m, p := newTestModel(t, 0, "a.ogg", "b.ogg", "c.ogg")
	press(m, "down")
	press(m, "down")
	press(m, "enter")
	if lastLoaded(p) != "/music/c.ogg" {
		t.Fatalf("loaded %q", p.loaded)
	}
}

func TestVolumeKeys(t *testing.T) {
	m, p := newTestModel(t, 0, "a.ogg")
	p.status.Volume = 0.5
	m.Update(tickMsg{})
	press(m, "+")
	press(m, "+")
	press(m, "-")
	if v := p.status.Volume; v < 0.549 || v > 0.551 {
		t.Fatalf("volume %v, want 0.55", v)
	}
}

func TestViewShowsTrackAndTimes(t *testing.T) {
	m, p := newTestModel(t, 0, "01 Song.ogg", "02 Other.ogg")
	p.status = audio.Status{Position: 65 * time.Second, Duration: 3 * time.Minute, Volume: 0.8}
	m.Update(tickMsg{})
	view := m.View()
	for _, want := range []string{"01 Song", "1:05 / 3:00", "vol 80%", "02 Other.ogg", "quit"} {
		if !strings.Contains(view, want) {
			t.Errorf("view lacks %q:\n%s", want, view)
		}
	}
	if lines := strings.Count(view, "\n") + 1; lines != 20-1 {
		t.Errorf("view has %d lines, want 19 (no error line)", lines)
	}
}

func TestFormatTime(t *testing.T) {
	for d, want := range map[time.Duration]string{
		0: "0:00", 59 * time.Second: "0:59", 61 * time.Second: "1:01", 3725 * time.Second: "1:02:05",
	} {
		if got := formatTime(d); got != want {
			t.Errorf("formatTime(%v) = %q, want %q", d, got, want)
		}
	}
}
