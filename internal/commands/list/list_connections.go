package list

import (
	"fmt"
	sshconn "term_cli/internal/ssh_conn"

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

			for idx, con := range *connections {
				fmt.Printf("[%d] %s@%s:%d\n", idx, con.GetProfile().Username, con.Host, con.Port)
			}

			return nil
		},
	}

	return listConnections
}
