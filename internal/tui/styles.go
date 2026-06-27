package tui

import "github.com/charmbracelet/lipgloss"

type Styles struct {
	theme      *Theme
	Title      lipgloss.Style
	Danger     lipgloss.Style
	Warning    lipgloss.Style
	Muted      lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Permanent  lipgloss.Style
	Trash      lipgloss.Style
	TrashPath  lipgloss.Style
	Border     lipgloss.Style
	KeyHint    lipgloss.Style
}

func NewStyles(theme *Theme) *Styles {
	if theme == nil {
		theme = DefaultTheme()
	}

	return &Styles{
		theme: theme,
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Colors.TitleFG)).
			Padding(0, 1),

		Danger: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Colors.DangerFG)),

		Warning: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Colors.WarningFG)),

		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Colors.MutedFG)),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Colors.SelectedFG)),

		Unselected: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Colors.UnselectedFG)),

		Permanent: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Colors.PermanentFG)).
			Background(lipgloss.Color(theme.Colors.PermanentBG)).
			Padding(0, 1),

		Trash: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(theme.Colors.TrashFG)).
			Background(lipgloss.Color(theme.Colors.TrashBG)).
			Padding(0, 1),

		TrashPath: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Colors.TrashPathFG)).
			Padding(0, 1),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(theme.Colors.BorderColor)).
			Padding(1, 2),

		KeyHint: lipgloss.NewStyle().
			Foreground(lipgloss.Color(theme.Colors.KeyHintFG)).
			Padding(0, 1),
	}
}

func DefaultStyles() *Styles {
	return NewStyles(ActiveTheme())
}

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
