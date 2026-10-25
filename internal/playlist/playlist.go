// Package playlist holds the ordered list of tracks and the rules of moving
// through it. It knows nothing about audio playback or the terminal.
package playlist

import "errors"

var ErrEmpty = errors.New("playlist: no audio files")

// Direction is where the listener last moved, so a track that fails to play
// is skipped the same way.
type Direction int

const (
	Forward  Direction = 1
	Backward Direction = -1
)

type Playlist struct {
	tracks  []Track
	current int
	dir     Direction
}

// New returns a playlist positioned at start.
func New(tracks []Track, start int) (*Playlist, error) {
	if len(tracks) == 0 {
		return nil, ErrEmpty
	}
	if start < 0 || start >= len(tracks) {
		start = 0
	}
	return &Playlist{tracks: tracks, current: start, dir: Forward}, nil
}

func (p *Playlist) Tracks() []Track { return p.tracks }

func (p *Playlist) Len() int { return len(p.tracks) }

func (p *Playlist) Index() int { return p.current }

func (p *Playlist) Current() Track { return p.tracks[p.current] }

// Next moves to the following track. It reports false at the end.
func (p *Playlist) Next() bool { return p.move(Forward) }

// Prev moves to the preceding track. It reports false at the start.
func (p *Playlist) Prev() bool { return p.move(Backward) }

// Skip moves past the current track in the last direction of travel.
func (p *Playlist) Skip() bool { return p.move(p.dir) }

// Select jumps to track i. Skipping from there goes forward.
func (p *Playlist) Select(i int) bool {
	if i < 0 || i >= len(p.tracks) {
		return false
	}
	p.current, p.dir = i, Forward
	return true
}

func (p *Playlist) move(d Direction) bool {
	i := p.current + int(d)
	if i < 0 || i >= len(p.tracks) {
		return false
	}
	p.current, p.dir = i, d
	return true
}
