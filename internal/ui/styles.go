package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

// ANSI colors, so the player follows the terminal theme.
var (
	accentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	cursorStyle = lipgloss.NewStyle().Bold(true)
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
)

func newHelp() help.Model {
	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle()
	h.Styles.ShortDesc = dimStyle
	h.Styles.ShortSeparator = dimStyle
	h.Styles.Ellipsis = dimStyle
	return h
}
