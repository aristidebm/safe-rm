package tui

import "example.com/safe-rm/internal/engine"

type ConfirmItem struct {
	Path     string
	IsDir    bool
	Policy   engine.Policy
	Expanded bool
	Children []string
	Selected bool
}

func RunConfirm(items []ConfirmItem) ([]ConfirmItem, bool) {
	return items, false
}
