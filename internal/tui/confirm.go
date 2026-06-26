package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"example.com/safe-rm/internal/engine"

	tea "github.com/charmbracelet/bubbletea"
)

type ConfirmItem struct {
	Path     string
	IsDir    bool
	Policy   engine.Policy
	Expanded bool
	Children []string
	Selected bool
	depth    int
}

type confirmModel struct {
	items    []ConfirmItem
	cursor   int
	height   int
	confirmed bool
	aborted  bool
}

func RunConfirm(items []ConfirmItem) ([]ConfirmItem, bool) {
	for i := range items {
		items[i].depth = 0
	}

	m := confirmModel{
		items:  items,
		cursor: 0,
	}

	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		return items, false
	}

	return items, m.confirmed
}

func (m *confirmModel) Init() tea.Cmd {
	return nil
}

func (m *confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.aborted = true
			return m, tea.Quit

		case "x":
			m.confirmed = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.visibleRows())-1 {
				m.cursor++
			}
			return m, nil

		case " ":
			rows := m.visibleRows()
			if m.cursor < len(rows) {
				idx := rows[m.cursor]
				m.items[idx].Selected = !m.items[idx].Selected
				if m.items[idx].IsDir && m.items[idx].Expanded {
					m.toggleChildren(idx, m.items[idx].Selected)
				}
			}
			return m, nil

		case "enter":
			rows := m.visibleRows()
			if m.cursor < len(rows) {
				idx := rows[m.cursor]
				if m.items[idx].IsDir {
					m.items[idx].Expanded = !m.items[idx].Expanded
					if m.items[idx].Expanded && len(m.items[idx].Children) == 0 {
						m.loadChildren(idx)
					}
				}
			}
			return m, nil

		case "a":
			for i := range m.items {
				if m.items[i].depth == 0 {
					m.items[i].Selected = true
				}
			}
			return m, nil

		case "n":
			for i := range m.items {
				m.items[i].Selected = false
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *confirmModel) loadChildren(parentIdx int) {
	entries, err := os.ReadDir(m.items[parentIdx].Path)
	if err != nil {
		return
	}

	insertIdx := parentIdx + 1
	for _, entry := range entries {
		child := ConfirmItem{
			Path:     filepath.Join(m.items[parentIdx].Path, entry.Name()),
			IsDir:    entry.IsDir(),
			Policy:   m.items[parentIdx].Policy,
			Selected: m.items[parentIdx].Selected,
			depth:    m.items[parentIdx].depth + 1,
		}

		m.items = append(m.items[:insertIdx], append([]ConfirmItem{child}, m.items[insertIdx:]...)...)
		insertIdx++
	}

	m.items[parentIdx].Children = make([]string, len(entries))
	for i, entry := range entries {
		m.items[parentIdx].Children[i] = entry.Name()
	}
}

func (m *confirmModel) toggleChildren(parentIdx int, selected bool) {
	for i := parentIdx + 1; i < len(m.items); i++ {
		if m.items[i].depth <= m.items[parentIdx].depth {
			break
		}
		m.items[i].Selected = selected
	}
}

func (m *confirmModel) visibleRows() []int {
	var rows []int
	for i, item := range m.items {
		if item.depth == 0 {
			rows = append(rows, i)
			if item.Expanded {
				for j := i + 1; j < len(m.items); j++ {
					if m.items[j].depth <= item.depth {
						break
					}
					rows = append(rows, j)
				}
			}
		}
	}
	return rows
}

func (m *confirmModel) View() string {
	var b strings.Builder

	b.WriteString(StyleBorder.Render(m.viewContent()))

	return b.String()
}

func (m *confirmModel) viewContent() string {
	var b strings.Builder

	b.WriteString(StyleTitle.Render("⚠  DANGER — Review files before deleting"))
	b.WriteString("\n\n")

	rows := m.visibleRows()
	selectedCount := 0
	permCount := 0

	for i, idx := range rows {
		item := m.items[idx]
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}

		checkbox := "[ ]"
		if item.Selected {
			checkbox = StyleSelected.Render("[✓]")
			selectedCount++
		}

		indent := strings.Repeat("  ", item.depth)
		prefix := ""
		if item.depth > 0 {
			if i < len(rows)-1 && m.items[rows[i+1]].depth >= item.depth {
				prefix = "├── "
			} else {
				prefix = "└── "
			}
		}

		icon := "📄"
		if item.IsDir {
			icon = "📁"
		}

		badge := ""
		switch item.Policy {
		case engine.PolicyDanger:
			badge = StyleTrash.Render(" TRASH ")
		case engine.PolicyDangerPermanent:
			badge = StylePermanent.Render(" PERMANENT ")
			permCount++
		}

		path := item.Path
		home, _ := os.UserHomeDir()
		if strings.HasPrefix(path, home) {
			path = "~" + path[len(home):]
		}

		line := fmt.Sprintf("%s%s%s%s %s %s %s\n",
			cursor, indent, prefix, checkbox, icon, path, badge)

		b.WriteString(line)
	}

	b.WriteString("\n")
	b.WriteString(StyleMuted.Render(fmt.Sprintf("%d selected · %d permanent · %d to trash",
		selectedCount, permCount, selectedCount-permCount)))
	b.WriteString("\n")

	b.WriteString(StyleKeyHint.Render("space: toggle  enter: expand dir  a: all  n: none"))
	b.WriteString("\n")
	b.WriteString(StyleKeyHint.Render("q/esc: abort   x: confirm deletion"))

	return b.String()
}
