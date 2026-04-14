package list

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/smsimone/sshm/internal/commands/add"
	sshconn "github.com/smsimone/sshm/internal/ssh_conn"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var listConnections = &cobra.Command{
		Use:   "list-connections",
		Short: "List registered connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			connections, err := sshconn.LoadConnections()
			if err != nil {
				return err
			}

			m := initModel(*connections)
			res, err := tea.NewProgram(m).Run()
			if err != nil {
				return fmt.Errorf("failed to list available connections")
			}
			if converted, ok := res.(*add.Model); ok && converted.Submitting {
				newConn := converted.GetConnection()
				return sshconn.UpdateConnection(newConn)
			}

			return nil
		},
	}

	return listConnections
}
