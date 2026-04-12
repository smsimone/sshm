package connect

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"

	sshconn "github.com/smsimone/sshm/internal/ssh_conn"

	tea "charm.land/bubbletea/v2"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func NewCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "connect [label]",
		Short: "Connect to a given host",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {

			connections, err := sshconn.LoadConnections()
			if err != nil {
				return err
			}

			m := newModel(*connections)
			res, err := tea.NewProgram(m).Run()
			if err != nil {
				return fmt.Errorf("failed to select connection: %w", err)
			} else if m.quitting {
				return nil
			}

			parsed := (res.(*model))
			conn := (*connections)[parsed.selected]

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

	if oldState, err := term.MakeRaw(fd); err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	} else {
		defer term.Restore(fd, oldState)
	}

	if err := session.RequestPty("xterm-256color", height, width, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
		ssh.ICANON:        1, // canonical input processing
		ssh.ISIG:          1, // enable signals
		ssh.IEXTEN:        1, // extended processing
	}); err != nil {
		return err
	}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh)
		defer signal.Stop(sigCh)
		for range sigCh {
			if w, h, err := term.GetSize(fd); err == nil {
				session.WindowChange(h, w)
			}
		}
	}()

	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	session.Stderr = os.Stderr

	if err := session.Shell(); err != nil {
		return err
	}

	done := make(chan error, 2)
	go func() {
		_, err := io.Copy(stdin, os.Stdin)
		stdin.Close()
		done <- err
	}()
	go func() {
		_, err := io.Copy(os.Stdout, stdout)
		done <- err
	}()

	<-done

	return nil
}
