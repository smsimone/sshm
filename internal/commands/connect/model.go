package connect

import (
	sshconn "github.com/smsimone/sshm/internal/ssh_conn"
	"github.com/smsimone/sshm/internal/styles"

	tea "charm.land/bubbletea/v2"
)

type model struct {
	connections []sshconn.Connection
	selected    int
	quitting    bool
}

var _ tea.Model = (*model)(nil)

func newModel(conns []sshconn.Connection) *model {
	return &model{
		connections: conns,
		selected:    0,
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

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
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			return m, tea.Quit
		}

	}
	return m, nil
}

func (m *model) View() tea.View {
	rows := []string{}
	for _, conn := range m.connections {
		rows = append(rows, conn.String(false))
	}

	body := styles.RenderList(styles.ListConfig{
		Title:      "Connections",
		RowContent: rows,
		Selected:   m.selected,
		Commands: []styles.CommandHelp{
			{Key: "q", Action: "quit"},
			{Key: "j/k", Action: "movement"},
			{Key: "return", Action: "connect"},
		},
	})

	return tea.NewView(body)
}
