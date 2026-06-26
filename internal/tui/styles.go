package tui

import "github.com/charmbracelet/lipgloss"

var (
	StyleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a2e", Dark: "#e0e0ff"}).
			Padding(0, 1)

	StyleDanger = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#ff4444"})

	StyleWarning = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#cc8800", Dark: "#ffcc00"})

	StyleMuted = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#666666"})

	StyleSelected = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#008800", Dark: "#44ff44"})

	StyleUnselected = lipgloss.NewStyle().
			 Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#555555"})

	StylePermanent = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#cc0000", Dark: "#ff4444"}).
			Background(lipgloss.AdaptiveColor{Light: "#ffcccc", Dark: "#440000"}).
			Padding(0, 1)

	StyleTrash = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "#0066cc", Dark: "#44aaff"}).
			Background(lipgloss.AdaptiveColor{Light: "#cce5ff", Dark: "#002244"}).
			Padding(0, 1)

	StyleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)

	StyleKeyHint = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#999999"}).
			Padding(0, 1)
)
