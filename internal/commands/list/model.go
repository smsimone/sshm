package list

import (
	tea "charm.land/bubbletea/v2"
	"github.com/smsimone/sshm/internal/commands/add"
	sshconn "github.com/smsimone/sshm/internal/ssh_conn"
	"github.com/smsimone/sshm/internal/styles"
)

type model struct {
	connections []sshconn.Connection
	editing     *sshconn.Connection
	selected    int
}

var _ tea.Model = (*model)(nil)

func initModel(connections []sshconn.Connection) *model {
	return &model{connections: connections, selected: 0}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func clamp(x, min, max int) int {
	if x <= min {
		return min
	} else if x >= max {
		return max
	}
	return x
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "ctrl+q", "esc":
			if m.editing != nil {
				m.editing = nil
				return m, nil
			}
			return m, tea.Quit
		case "j", "k":
			if msg.String() == "j" {
				m.selected += 1
			} else {
				m.selected -= 1
			}
			m.selected = clamp(m.selected, 0, len(m.connections))
			return m, nil
		case "e":
			// m.editing = &m.connections[m.selected]
			newM := add.InitialModel(&m.connections[m.selected])
			return newM, nil
		}
	}

	return m, nil
}

func (m *model) View() tea.View {
	lines := make([]string, len(m.connections))
	for idx, conn := range m.connections {
		lines[idx] = conn.String(false)
	}

	content := styles.RenderList(styles.ListConfig{
		Title:      "Connections",
		Selected:   m.selected,
		RowContent: lines,
	})

	return tea.NewView(content)
}
