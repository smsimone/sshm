package profile

import (
	sshconn "sshm/internal/ssh_conn"
	"sshm/internal/styles"

	tea "charm.land/bubbletea/v2"
)

type listProfileModel struct {
	profiles []sshconn.Profile
	selected int
}

var _ tea.Model = (*listProfileModel)(nil)

func initialListProfileModel() *listProfileModel {
	profiles, err := sshconn.LoadProfiles()
	if err != nil {
		panic(err)
	}

	return &listProfileModel{profiles: *profiles}
}

func (l *listProfileModel) Init() tea.Cmd {
	return nil
}

// Update implements [tea.Model].
func (l *listProfileModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "ctrl+c", "q":
			return l, tea.Quit
		case "j":
			if l.selected+1 < len(l.profiles) {
				l.selected += 1
			}
			return l, tea.ClearScreen
		case "k":
			if l.selected > 0 {
				l.selected -= 1
			}
			return l, tea.ClearScreen
		case "d", "D":
			force := false
			if msg.String() == "D" {
				force = true
			}
			if err := sshconn.RemoveProfile(l.profiles[l.selected].Label, force); err != nil {
				return l, tea.Printf("%s", err.Error())
			}
			l.updateModel()
		}
	}

	return l, nil
}

func (l *listProfileModel) updateModel() {
	profiles, err := sshconn.LoadProfiles()
	if err != nil {
		panic(err)
	}
	l.profiles = *profiles
}

// View implements [tea.Model].
func (l *listProfileModel) View() tea.View {

	rows := []string{}
	for _, profile := range l.profiles {
		rows = append(rows, profile.Label)
	}

	body := styles.RenderList(styles.ListConfig{
		Title:      "Profiles",
		RowContent: rows,
		Selected:   l.selected,
		Commands: []styles.CommandHelp{
			{Key: "q", Action: "quit"},
			{Key: "j/k", Action: "movement"},
			{Key: "d/D", Action: "delete"},
		},
	})

	return tea.NewView(body)
}
