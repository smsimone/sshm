package connect

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"strings"
	"syscall"
	appstate "term_cli/internal/handlers"
	sshconn "term_cli/internal/ssh_conn"
	"time"

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
			if conn, err := appstate.LoadFile(); err != nil {
				return err
			} else {
				conns = *conn
			}

			if len(args) == 0 {

				for idx, con := range conns {
					fmt.Printf("[%d] (%s) %s@%s:%d\n", idx, con.Label, con.Username, con.Host, con.Port)
				}
				fmt.Printf("Which connection do you want? ")
				reader := bufio.NewReader(os.Stdin)
				line, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read line: %s", err.Error())
				}
				line = strings.TrimSpace(line)
				idx, err := strconv.Atoi(line)
				if err != nil {
					return fmt.Errorf("invalid idx: %s", line)
				}
				connection = conns[idx]
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
