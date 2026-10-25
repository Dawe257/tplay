package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/x/ansi"
)

// Rows besides the playlist: title, directory, blank, progress, blank,
// playlist header, help and error.
const chromeRows = 8

func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}
	width := m.width - 2 // one column of padding on each side
	var lines []string
	add := func(s string) { lines = append(lines, " "+ansi.Truncate(s, width, "…")) }

	track := m.playlist.Current()
	add(accentStyle.Render(m.stateIcon() + "  " + track.Title()))
	add(dimStyle.Render(track.Dir()))
	add("")
	add(m.progressLine(width))
	add("")
	add(dimStyle.Render(fmt.Sprintf("Playlist %d/%d", m.playlist.Index()+1, m.playlist.Len())))
	for _, row := range m.playlistRows(max(m.height-chromeRows, 1)) {
		add(row)
	}
	add(m.help.View(m.keys))
	if m.err != "" {
		add(errStyle.Render(m.err))
	}
	return strings.Join(lines, "\n")
}

func (m *Model) stateIcon() string {
	switch {
	case m.ended:
		return "■"
	case m.status.Paused:
		return "⏸"
	}
	return "▶"
}

func (m *Model) progressLine(width int) string {
	pos, total := m.status.Position, m.status.Duration
	if m.ended {
		pos = total
	}
	info := fmt.Sprintf(" %s / %s  vol %d%%", formatTime(pos), formatTime(total),
		int(m.status.Volume*100+0.5))
	return progressBar(pos, total, width-ansi.StringWidth(info)) + dimStyle.Render(info)
}

// playlistRows renders a window of the playlist that keeps the cursor visible,
// padded to exactly rows lines so the help stays at the bottom.
func (m *Model) playlistRows(rows int) []string {
	tracks := m.playlist.Tracks()
	first := min(max(m.cursor-rows/2, 0), max(len(tracks)-rows, 0))
	out := make([]string, 0, rows)
	for i := first; i < min(first+rows, len(tracks)); i++ {
		name := fmt.Sprintf("%3d  %s", i+1, tracks[i].FileName())
		cursor, playing := " ", " "
		if i == m.playlist.Index() {
			playing = "▶"
			name = accentStyle.Render(name)
		}
		if i == m.cursor {
			cursor = cursorStyle.Render("›")
			if i != m.playlist.Index() {
				name = cursorStyle.Render(name)
			}
		}
		out = append(out, cursor+playing+name)
	}
	for len(out) < rows {
		out = append(out, "")
	}
	return out
}

func progressBar(pos, total time.Duration, width int) string {
	if width <= 0 {
		return ""
	}
	filled := 0
	if total > 0 {
		filled = min(int(float64(width)*float64(pos)/float64(total)), width)
	}
	return accentStyle.Render(strings.Repeat("━", filled)) + dimStyle.Render(strings.Repeat("─", width-filled))
}

func formatTime(d time.Duration) string {
	s := int(d.Round(time.Second) / time.Second)
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s/60%60, s%60)
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}
