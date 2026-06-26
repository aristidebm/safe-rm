package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"example.com/safe-rm/internal/engine"

	tea "github.com/charmbracelet/bubbletea"
)

type RestoreResult struct {
	ToRestore []*engine.TrashEntry
	ToDelete  []*engine.TrashEntry
	Conflict  engine.ConflictStrategy
	Aborted   bool
}

type restoreModel struct {
	entries   []*engine.TrashEntry
	selected  map[int]bool
	cursor    int
	filter    string
	filtering bool
	conflict  bool
	conflictPath string
	result    *RestoreResult
	height    int
}

func RunRestore(entries []*engine.TrashEntry) *RestoreResult {
	m := restoreModel{
		entries:  entries,
		selected: make(map[int]bool),
		result:   &RestoreResult{Aborted: true},
	}

	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		return &RestoreResult{Aborted: true}
	}

	return m.result
}

func (m *restoreModel) Init() tea.Cmd {
	return nil
}

func (m *restoreModel) filteredEntries() []int {
	var indices []int
	for i, e := range m.entries {
		if m.filter == "" || strings.Contains(strings.ToLower(e.OriginalPath), strings.ToLower(m.filter)) {
			indices = append(indices, i)
		}
	}
	return indices
}

func (m *restoreModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.conflict {
			return m, m.handleConflictKey(msg)
		}

		if m.filtering {
			return m, m.handleFilterKey(msg)
		}

		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit

		case "up", "k":
			filtered := m.filteredEntries()
			if m.cursor > 0 {
				m.cursor--
			}
			_ = filtered
			return m, nil

		case "down", "j":
			filtered := m.filteredEntries()
			if m.cursor < len(filtered)-1 {
				m.cursor++
			}
			return m, nil

		case " ":
			filtered := m.filteredEntries()
			if m.cursor < len(filtered) {
				idx := filtered[m.cursor]
				m.selected[idx] = !m.selected[idx]
			}
			return m, nil

		case "/":
			m.filtering = true
			m.filter = ""
			return m, nil

		case "a":
			filtered := m.filteredEntries()
			for _, idx := range filtered {
				m.selected[idx] = true
			}
			return m, nil

		case "n":
			for k := range m.selected {
				delete(m.selected, k)
			}
			return m, nil

		case "r":
			return m, m.restoreSelected()

		case "d":
			return m, m.deleteSelected()
		}
	}

	return m, nil
}

func (m *restoreModel) handleFilterKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "enter":
		m.filtering = false
		return nil

	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.cursor = 0
		}
		return nil

	default:
		if len(msg.String()) == 1 && msg.String()[0] >= 32 {
			m.filter += msg.String()
			m.cursor = 0
		}
		return nil
	}
}

func (m *restoreModel) handleConflictKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "r":
		m.result.Conflict = engine.ConflictRename
		m.result.Aborted = false
		return tea.Quit
	case "o":
		m.result.Conflict = engine.ConflictOverwrite
		m.result.Aborted = false
		return tea.Quit
	case "s":
		m.result.Conflict = engine.ConflictSkip
		m.result.Aborted = false
		return tea.Quit
	}
	return nil
}

func (m *restoreModel) restoreSelected() tea.Cmd {
	var toRestore []*engine.TrashEntry
	for i, e := range m.entries {
		if m.selected[i] {
			toRestore = append(toRestore, e)
		}
	}

	if len(toRestore) == 0 {
		return nil
	}

	for _, e := range toRestore {
		if _, err := os.Stat(e.OriginalPath); err == nil {
			m.conflict = true
			m.conflictPath = e.OriginalPath
			m.result.ToRestore = toRestore
			return nil
		}
	}

	m.result.ToRestore = toRestore
	m.result.Aborted = false
	return tea.Quit
}

func (m *restoreModel) deleteSelected() tea.Cmd {
	var toDelete []*engine.TrashEntry
	for i, e := range m.entries {
		if m.selected[i] {
			toDelete = append(toDelete, e)
		}
	}

	if len(toDelete) == 0 {
		return nil
	}

	m.result.ToDelete = toDelete
	m.result.Aborted = false
	return tea.Quit
}

func (m *restoreModel) View() string {
	var b strings.Builder

	b.WriteString(StyleBorder.Render(m.viewContent()))
	return b.String()
}

func (m *restoreModel) viewContent() string {
	var b strings.Builder

	b.WriteString(StyleTitle.Render("🗑  Trash — select files to restore"))
	b.WriteString("\n\n")

	filterLabel := ""
	if m.filtering {
		filterLabel = StyleWarning.Render(" Filter: " + m.filter + "▎ ")
	} else if m.filter != "" {
		filterLabel = StyleMuted.Render(" Filter: " + m.filter + " ")
	}
	if filterLabel != "" {
		b.WriteString(filterLabel)
		b.WriteString("\n\n")
	}

	filtered := m.filteredEntries()
	selectedCount := 0

	for displayIdx, origIdx := range filtered {
		e := m.entries[origIdx]
		cursor := "  "
		if displayIdx == m.cursor {
			cursor = "▸ "
		}

		checkbox := "[ ]"
		if m.selected[origIdx] {
			checkbox = StyleSelected.Render("[✓]")
			selectedCount++
		}

		icon := "📄"
		if e.IsDir {
			icon = "📁"
		}

		path := e.OriginalPath
		home, _ := os.UserHomeDir()
		if strings.HasPrefix(path, home) {
			path = "~" + path[len(home):]
		}

		trashPath := filepath.Join("Trash", "files", e.ID)
		if _, err := os.Stat(trashPath); os.IsNotExist(err) {
			path += StyleDanger.Render(" [missing]")
		}

		sizeStr := formatSize(e.Size)
		timeStr := e.TrashedAt.Format("2006-01-02 15:04:05")

		line := fmt.Sprintf("%s%s %s %s %s  %s\n",
			cursor, checkbox, icon, path, sizeStr, timeStr)

		b.WriteString(line)
	}

	b.WriteString("\n")
	b.WriteString(StyleMuted.Render(fmt.Sprintf("%d selected", selectedCount)))
	b.WriteString("\n")

	if m.conflict {
		b.WriteString(StyleWarning.Render(fmt.Sprintf("Conflict: %s already exists", m.conflictPath)))
		b.WriteString("\n")
		b.WriteString(StyleKeyHint.Render("[R]ename   [O]verwrite   [S]kip"))
	} else {
		b.WriteString(StyleKeyHint.Render("/: filter  space: toggle  a: all  n: none"))
		b.WriteString("\n")
		b.WriteString(StyleKeyHint.Render("r: restore selected  d: delete selected  q: abort"))
	}

	return b.String()
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}


