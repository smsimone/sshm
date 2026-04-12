package add

import (
	"fmt"

	sshconn "github.com/smsimone/sshm/internal/ssh_conn"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

func NewTeaCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add-connection",
		Short: "Add a new connection to the file",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := initialModel()
			res, err := tea.NewProgram(p).Run()
			if err != nil {
				return err
			} else if p.quitting {
				return nil
			}
			model := res.(model)
			return sshconn.AddItem(model.GetConnection())
		},
	}
	return cmd
}

func NewCommand() *cobra.Command {
	var (
		label       string
		host        string
		port        int
		profileName string
	)

	var cmd = &cobra.Command{
		Use:   "add-connection",
		Short: "Add a new connection to the file",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			_, err := sshconn.GetProfile(label)
			if err != nil {
				return fmt.Errorf("failed to load configuration file: %s", err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := sshconn.Connection{
				Label:   label,
				Host:    host,
				Port:    port,
				Profile: &profileName,
			}

			return sshconn.AddItem(conn)
		},
	}

	cmd.Flags().StringVarP(&label, "label", "l", "", "Label for the connection")
	if err := cmd.MarkFlagRequired("label"); err != nil {
		panic(err)
	}
	cmd.Flags().StringVarP(&host, "host", "H", "", "Address of the host")
	if err := cmd.MarkFlagRequired("host"); err != nil {
		panic(err)
	}
	cmd.Flags().StringVarP(&profileName, "profile", "p", "", "Profile to use to connect to host")

	cmd.Flags().IntVarP(&port, "port", "P", 22, "Port which handles ssh connection")

	return cmd
}
