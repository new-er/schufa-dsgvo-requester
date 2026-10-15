package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Color palette (exported for other packages)
	Subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	Highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	Text      = lipgloss.AdaptiveColor{Light: "#1A1A2E", Dark: "#DDDDDD"}

	// Tab styles
	TabStyle = lipgloss.NewStyle().
		Foreground(Subtle).
		Background(lipgloss.Color("0")).
		Padding(0, 2).
		MarginRight(1)

	ActiveTabStyle = TabStyle.Copy().
		Foreground(Text).
		Background(Highlight).
		Bold(true).
		Underline(true)

	// Content pane style
	ContentStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Highlight).
		Padding(1, 2).
		Foreground(Text)
)