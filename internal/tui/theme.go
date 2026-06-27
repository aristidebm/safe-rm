package tui

var trashPath string

func SetTrashPath(path string) {
	trashPath = path
}

func TrashPath() string {
	return trashPath
}

type ThemeColors struct {
	TitleFG      string `toml:"title_fg"`
	DangerFG     string `toml:"danger_fg"`
	WarningFG    string `toml:"warning_fg"`
	MutedFG      string `toml:"muted_fg"`
	SelectedFG   string `toml:"selected_fg"`
	UnselectedFG string `toml:"unselected_fg"`
	PermanentFG  string `toml:"permanent_fg"`
	PermanentBG  string `toml:"permanent_bg"`
	TrashFG      string `toml:"trash_fg"`
	TrashBG      string `toml:"trash_bg"`
	TrashPathFG  string `toml:"trash_path_fg"`
	BorderColor  string `toml:"border_color"`
	KeyHintFG    string `toml:"keyhint_fg"`
}

type Theme struct {
	Name   string      `toml:"name"`
	Author string      `toml:"author"`
	Colors ThemeColors `toml:"colors"`
}

var activeTheme *Theme

func SetTheme(t *Theme) {
	activeTheme = t
}

func ActiveTheme() *Theme {
	if activeTheme != nil {
		return activeTheme
	}
	return DefaultTheme()
}

func DefaultTheme() *Theme {
	return &Theme{
		Name:   "Default",
		Author: "safe-rm",
		Colors: ThemeColors{
			TitleFG:      "#1a1a2e",
			DangerFG:     "#cc0000",
			WarningFG:    "#cc8800",
			MutedFG:      "#888888",
			SelectedFG:   "#008800",
			UnselectedFG: "#888888",
			PermanentFG:  "#cc0000",
			PermanentBG:  "#ffcccc",
			TrashFG:      "#0066cc",
			TrashBG:      "#cce5ff",
			TrashPathFG:  "#aaaaaa",
			BorderColor:  "#888888",
			KeyHintFG:    "#666666",
		},
	}
}

func DarkTheme() *Theme {
	return &Theme{
		Name:   "Dark",
		Author: "safe-rm",
		Colors: ThemeColors{
			TitleFG:      "#e0e0ff",
			DangerFG:     "#ff4444",
			WarningFG:    "#ffcc00",
			MutedFG:      "#666666",
			SelectedFG:   "#44ff44",
			UnselectedFG: "#555555",
			PermanentFG:  "#ff4444",
			PermanentBG:  "#440000",
			TrashFG:      "#44aaff",
			TrashBG:      "#002244",
			TrashPathFG:  "#666666",
			BorderColor:  "#555555",
			KeyHintFG:    "#999999",
		},
	}
}
