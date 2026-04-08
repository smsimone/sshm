package connect

import (
	"fmt"
	sshconn "sshm/internal/ssh_conn"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	connections []sshconn.Connection
	selected    int
}

var _ tea.Model = (*model)(nil)

// Init implements [tea.Model].
func (m *model) Init() tea.Cmd {
	return nil
}

// Update implements [tea.Model].
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "j":
			if m.selected+1 < len(m.connections) {
				m.selected += 1
			}
		case "k":
			if m.selected > 0 {
				m.selected -= 1
			}
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		}

	}
	return m, nil
}

// View implements [tea.Model].
func (m *model) View() tea.View {
	var msg strings.Builder

	for idx := range m.connections {
		conn := m.connections[idx]
		if idx == m.selected {
			fmt.Fprintf(&msg, "> %s\n", conn.String(false))
		} else {
			fmt.Fprintf(&msg, "  %s\n", conn.String(false))
		}
	}

	return tea.NewView(msg.String())
}
