package list

import (
	"fmt"
	appstate "term_cli/internal/handlers"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var listConnections = &cobra.Command{
		Use:   "list-connections",
		Short: "List registered connections",
		RunE: func(cmd *cobra.Command, args []string) error {
			connections, err := appstate.LoadFile()
			if err != nil {
				return err
			}

			for idx, con := range *connections {
				fmt.Printf("[%d] %s@%s:%d\n", idx, con.Username, con.Host, con.Port)
			}

			return nil
		},
	}

	return listConnections
}
