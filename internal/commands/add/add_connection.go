package add

import (
	sshconn "github.com/smsimone/sshm/internal/ssh_conn"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add-connection",
		Short: "Add a new connection to the file",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := initialModel()
			res, err := tea.NewProgram(p).Run()
			if err != nil {
				return err
			} else if !p.submitting {
				return nil
			}
			model := res.(model)
			return sshconn.AddItem(model.GetConnection())
		},
	}
	return cmd
}
