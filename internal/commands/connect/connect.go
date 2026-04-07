package connect

import (
	"fmt"
	"net"
	"os"
	"slices"
	"strconv"
	appstate "term_cli/internal/handlers"
	sshconn "term_cli/internal/ssh_conn"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func NewCommand() *cobra.Command {
	var conns []sshconn.Connection

	cmd := &cobra.Command{
		Use:   "connect [label]",
		Short: "Connect to a given host",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]
			if conn, err := appstate.LoadFile(); err != nil {
				return err
			} else {
				conns = *conn
			}

			if !slices.ContainsFunc(conns, func(con sshconn.Connection) bool { return con.Label == label }) {
				return fmt.Errorf("Label %s not available", label)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			idx := slices.IndexFunc(conns, func(x sshconn.Connection) bool { return x.Label == label })
			conn := conns[idx]

			fmt.Printf("Connecting to %s:%d\n", conn.Host, conn.Port)

			client, err := ssh.Dial("tcp", net.JoinHostPort(conn.Host, strconv.FormatInt(int64(conn.Port), 10)), &ssh.ClientConfig{
				User:            conn.Username,
				Auth:            conn.AuthMethods(),
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			})
			if err != nil {
				return fmt.Errorf("failed to create ssh client: %w", err)
			}

			session, err := client.NewSession()
			if err != nil {
				return fmt.Errorf("failed to create new session: %w", err)
			}
			defer session.Close()

			session.Stdout = os.Stdout
			session.Stdin = os.Stdin
			session.Stderr = os.Stderr

			if err := session.Shell(); err != nil {
				return fmt.Errorf("failed to start shell: %s", err.Error())
			}

			return nil
		},
	}

	return cmd
}
