package connect

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"slices"
	sshconn "sshm/internal/ssh_conn"
	"strconv"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func NewCommand() *cobra.Command {
	var conns []sshconn.Connection
	var connection sshconn.Connection

	cmd := &cobra.Command{
		Use:   "connect [label]",
		Short: "Connect to a given host",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if conn, err := sshconn.LoadConnections(); err != nil {
				return err
			} else {
				conns = *conn
			}

			if len(args) == 0 {
				m := model{connections: conns, selected: 0}
				res, err := tea.NewProgram(&m).Run()
				if err != nil {
					return fmt.Errorf("failed to select connection: %w", err)
				}

				parsed := (res.(*model))

				connection = parsed.connections[parsed.selected]
				return nil
			} else {
				label := args[0]

				idx := slices.IndexFunc(conns, func(con sshconn.Connection) bool { return con.Label == label })
				if idx == -1 {
					return fmt.Errorf("Label %s not available", label)
				}
				connection = conns[idx]
				return nil
			}

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			defer func() {
				fmt.Printf("Connection terminated after %dms", time.Now().UnixMilli()-start.UnixMilli())
			}()

			conn := connection

			client, err := ssh.Dial("tcp", net.JoinHostPort(conn.Host, strconv.FormatInt(int64(conn.Port), 10)), &ssh.ClientConfig{
				User:            conn.GetProfile().Username,
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

			session.Stdout = crlfWriter{os.Stdout}
			session.Stderr = crlfWriter{os.Stderr}
			session.Stdin = os.Stdin

			if err := requestPty(session); err != nil {
				return fmt.Errorf("failed to request pty: %s", err.Error())
			}

			return session.Wait()
		},
	}

	return cmd
}

func requestPty(session *ssh.Session) error {
	width, height := 80, 24

	fd := int(os.Stdin.Fd())
	if w, h, err := term.GetSize(fd); err == nil {
		width = w
		height = h
	}

	err := session.RequestPty("xterm-256color", height, width, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	})
	if err != nil {
		return err
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer term.Restore(fd, oldState)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGWINCH)
		for range sigCh {
			if w, h, err := term.GetSize(fd); err == nil {
				session.WindowChange(h, w)
			}
		}
	}()

	if err := session.Shell(); err != nil {
		return err
	}

	return nil
}

type crlfWriter struct {
	w io.Writer
}

func (c crlfWriter) Write(b []byte) (int, error) {
	replaced := bytes.ReplaceAll(b, []byte{'\n'}, []byte{'\r', '\n'})
	_, err := c.w.Write(replaced)
	return len(b), err // ritorna len(b) originale, non quello replaced
}
