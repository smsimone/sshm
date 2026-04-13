package list

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
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
			if _, err = tea.NewProgram(m).Run(); err != nil {
				return fmt.Errorf("failed to list available connections")
			}

			return nil
		},
	}

	return listConnections
}
