// Package ui is the terminal interface. It drives a Player through the
// playlist and renders the state; it doesn't know how audio is produced.
package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"tplay/internal/audio"
	"tplay/internal/playlist"
)

const (
	tickInterval = 200 * time.Millisecond
	seekStep     = 5 * time.Second
	volumeStep   = 0.05
	// Prev restarts the current track instead when it played longer than this.
	restartAfter = 3 * time.Second
)

// Player is what the interface needs from the playback engine.
type Player interface {
	Load(path string)
	Stop()
	TogglePause()
	Seek(to time.Duration)
	SeekBy(delta time.Duration)
	SetVolume(v float64)
	Status() audio.Status
}

type tickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(tickInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

type Model struct {
	player   Player
	playlist *playlist.Playlist
	keys     keyMap
	help     help.Model

	status audio.Status
	cursor int
	ended  bool // the last track played to the end
	err    string
	width  int
	height int
}

func New(player Player, pl *playlist.Playlist) *Model {
	return &Model{player: player, playlist: pl, keys: defaultKeyMap(), help: newHelp()}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.load(), tick())
}

// load starts the current playlist track.
func (m *Model) load() tea.Cmd {
	track := m.playlist.Current()
	m.player.Load(track.Path)
	m.status = m.player.Status()
	m.cursor, m.ended = m.playlist.Index(), false
	return tea.SetWindowTitle("tplay - " + track.Title())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.Width = msg.Width - 2
		return m, nil
	case tickMsg:
		return m, tea.Batch(m.onTick(), tick())
	case tea.KeyMsg:
		return m, m.onKey(msg)
	}
	return m, nil
}

// onTick refreshes the status and moves on when the track is over.
func (m *Model) onTick() tea.Cmd {
	m.status = m.player.Status()
	if !m.status.Finished || m.ended {
		return nil
	}
	if err := m.status.Err; err != nil {
		m.err = fmt.Sprintf("%s: %v", m.playlist.Current().FileName(), err)
		if m.playlist.Skip() {
			return m.load()
		}
	} else if m.playlist.Next() {
		return m.load()
	}
	m.ended = true
	return nil
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	k := m.keys
	switch {
	case key.Matches(msg, k.Quit):
		m.player.Stop()
		return tea.Quit
	case key.Matches(msg, k.Pause):
		if m.ended {
			return m.load()
		}
		m.player.TogglePause()
	case key.Matches(msg, k.SeekBack):
		m.player.SeekBy(-seekStep)
	case key.Matches(msg, k.SeekFwd):
		m.player.SeekBy(seekStep)
	case key.Matches(msg, k.Next):
		if m.playlist.Next() {
			return m.load()
		}
	case key.Matches(msg, k.Prev):
		if m.status.Position <= restartAfter && m.playlist.Prev() {
			return m.load()
		}
		m.player.Seek(0)
	case key.Matches(msg, k.Up):
		m.cursor = max(m.cursor-1, 0)
	case key.Matches(msg, k.Down):
		m.cursor = min(m.cursor+1, m.playlist.Len()-1)
	case key.Matches(msg, k.Top):
		m.cursor = 0
	case key.Matches(msg, k.Bottom):
		m.cursor = m.playlist.Len() - 1
	case key.Matches(msg, k.Play):
		if m.playlist.Select(m.cursor) {
			return m.load()
		}
	case key.Matches(msg, k.VolUp):
		m.player.SetVolume(m.status.Volume + volumeStep)
	case key.Matches(msg, k.VolDown):
		m.player.SetVolume(m.status.Volume - volumeStep)
	}
	m.status = m.player.Status()
	return nil
}
