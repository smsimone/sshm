package styles

import (
	"fmt"

	"strings"

	"charm.land/lipgloss/v2"
)

var (
	ListTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			MarginBottom(1)

	ListItemStyle = lipgloss.NewStyle().
			Padding(0, 1)

	ListSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("99")).
				Padding(0, 1)

	ListHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			MarginTop(1)

	ListBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)
)

type CommandHelp struct {
	Key    string
	Action string
}

type ListConfig struct {
	Title      string
	RowContent []string
	Selected   int
	Commands   []CommandHelp
}

func RenderList(config ListConfig) string {
	var list strings.Builder
	var out strings.Builder

	out.WriteString(ListTitleStyle.Render(config.Title))
	out.WriteRune('\n')

	for idx := range config.RowContent {
		profile := config.RowContent[idx]
		if idx == config.Selected {
			fmt.Fprintln(&list, ListSelectedStyle.Render("> "+profile))
		} else {
			fmt.Fprintln(&list, ListItemStyle.Render("  "+profile))
		}
	}

	if len(config.RowContent) == 0 {
		fmt.Fprintln(&list, ListItemStyle.Render("(no content)"))
	}

	out.WriteString(list.String())
	out.WriteRune('\n')

	if len(config.Commands) > 0 {
		var commandsHelp strings.Builder
		for idx, cmd := range config.Commands {
			isLast := idx == len(config.Commands)
			fmt.Fprintf(&commandsHelp, "%s %s", cmd.Key, cmd.Action)
			if !isLast {
				fmt.Fprintf(&commandsHelp, " | ")
			}
		}
		out.WriteString(ListHelpStyle.Render(commandsHelp.String()))
	}

	return ListBoxStyle.Render(out.String())
}
