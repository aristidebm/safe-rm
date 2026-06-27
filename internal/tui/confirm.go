package tui

import (
	"fmt"
	"os"
	"strings"

	"example.com/safe-rm/internal/engine"

	tea "github.com/charmbracelet/bubbletea"
)

type confirmModel struct {
	root      *engine.Node
	nodes     []*engine.Node
	cursor    int
	height    int
	confirmed bool
	aborted   bool
	styles    *Styles
}

func RunConfirm(root *engine.Node) (*engine.Node, bool) {
	root.Expand()

	m := confirmModel{
		root:   root,
		nodes:  root.VisibleNodes(),
		cursor: 0,
		styles: DefaultStyles(),
	}

	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		return root, false
	}

	return root, m.confirmed
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
			if m.cursor < len(m.nodes)-1 {
				m.cursor++
			}
			return m, nil

		case " ":
			if m.cursor < len(m.nodes) {
				node := m.nodes[m.cursor]
				node.Selected = !node.Selected
				if node.IsDir && node.Expanded {
					toggleChildren(node, node.Selected)
				}
			}
			return m, nil

		case "enter":
			if m.cursor < len(m.nodes) {
				node := m.nodes[m.cursor]
				if node.IsDir {
					if node.Expanded {
						node.Collapse()
					} else {
						node.Expand()
					}
					m.nodes = m.root.VisibleNodes()
				}
			}
			return m, nil

		case "a":
			selectAll(m.root, true)
			m.nodes = m.root.VisibleNodes()
			return m, nil

		case "n":
			selectAll(m.root, false)
			m.nodes = m.root.VisibleNodes()
			return m, nil
		}
	}

	return m, nil
}

func toggleChildren(parent *engine.Node, selected bool) {
	for _, child := range parent.Children {
		child.Selected = selected
		if child.IsDir {
			toggleChildren(child, selected)
		}
	}
}

func selectAll(node *engine.Node, selected bool) {
	node.Selected = selected
	for _, child := range node.Children {
		selectAll(child, selected)
	}
}

func (m *confirmModel) View() string {
	var b strings.Builder

	b.WriteString(m.styles.Border.Render(m.viewContent()))

	return b.String()
}

func (m *confirmModel) viewContent() string {
	var b strings.Builder

	b.WriteString(m.styles.Title.Render("DANGER — Review files before deleting"))
	b.WriteString("\n\n")

	selectedCount := 0
	permCount := 0

	for i, node := range m.nodes {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		checkbox := "[ ]"
		if node.Selected {
			checkbox = m.styles.Selected.Render("[x]")
			selectedCount++
		}

		indent := strings.Repeat("  ", node.Depth)
		prefix := ""
		if node.Depth > 0 {
			isLast := false
			if i < len(m.nodes)-1 && m.nodes[i+1].Depth <= node.Depth {
				isLast = true
			}
			if isLast || i == len(m.nodes)-1 {
				prefix = "\\-- "
			} else {
				prefix = "|-- "
			}
		}

		icon := " "
		if node.IsDir {
			if node.Expanded {
				icon = "v"
			} else {
				icon = ">"
			}
		}

		badge := ""
		switch node.Policy {
		case engine.PolicyDanger:
			badge = m.styles.Trash.Render(" TRASH ")
		case engine.PolicyDangerPermanent:
			badge = m.styles.Permanent.Render(" PERMANENT ")
			permCount++
		}

		path := node.Path
		home, _ := os.UserHomeDir()
		if strings.HasPrefix(path, home) {
			path = "~" + path[len(home):]
		}

		line := fmt.Sprintf("%s%s%s%s %s %s %s\n",
			cursor, indent, prefix, checkbox, icon, path, badge)

		b.WriteString(line)
	}

	b.WriteString("\n")
	b.WriteString(m.styles.Muted.Render(fmt.Sprintf("%d selected  %d permanent  %d to trash",
		selectedCount, permCount, selectedCount-permCount)))
	b.WriteString("\n")

	b.WriteString(m.styles.KeyHint.Render("space: toggle  enter: expand/collapse dir  a: all  n: none"))
	b.WriteString("\n")
	b.WriteString(m.styles.KeyHint.Render("q/esc: abort   x: confirm deletion"))
	b.WriteString("\n")
	b.WriteString(m.styles.TrashPath.Render("trash: " + TrashPath()))

	return b.String()
}
