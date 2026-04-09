package profile

import (
	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

func ListProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-profiles",
		Short: "List the registered profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := initialListProfileModel()
			_, err := tea.NewProgram(p).Run()
			if err != nil {
				return err
			}

			return nil
		},
	}
	return cmd
}
