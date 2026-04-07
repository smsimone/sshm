package add

import (
	"encoding/base64"
	"fmt"
	"os"
	appstate "term_cli/internal/handlers"
	sshconn "term_cli/internal/ssh_conn"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func NewCommand() *cobra.Command {
	var (
		label         string
		host          string
		username      string
		port          int
		password      string
		keyPath       string
		keyPassphrase string
	)

	var cmd = &cobra.Command{
		Use:   "add-connection",
		Short: "Add a new connection to the file",
		RunE: func(cmd *cobra.Command, args []string) error {
			conn := sshconn.Connection{
				Label:    label,
				Host:     host,
				Port:     port,
				Username: username,
			}

			if len(keyPath) > 0 {
				pvtKey, err := loadPrivateKey(keyPath, &keyPassphrase)
				if err != nil {
					return err
				}
				conn.PvtKey = pvtKey
			}

			return appstate.AddItem(conn)
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
	cmd.Flags().StringVarP(&username, "username", "u", os.Getenv("USER"), "Username")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password")

	cmd.Flags().StringVarP(&keyPath, "keypath", "k", "", "Private key filepath")
	cmd.Flags().StringVarP(&keyPassphrase, "passphrase", "t", "", "Private key's passphrase")

	cmd.MarkFlagsMutuallyExclusive("password", "keypath")

	cmd.Flags().IntVarP(&port, "port", "P", 22, "Port which handles ssh connection")

	return cmd
}

func loadPrivateKey(keyPath string, keyPassphrase *string) (*sshconn.PrivateKey, error) {
	if stat, err := os.Stat(keyPath); err != nil {
		return nil, fmt.Errorf("private key does not exists")
	} else if stat.IsDir() {
		return nil, fmt.Errorf("private key path is pointing to a directory")
	}

	content, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	if _, err := ssh.ParsePrivateKey(content); err != nil {
		return nil, err
	}

	encodedContent := base64.StdEncoding.EncodeToString(content)

	pvt := sshconn.PrivateKey{
		Content:    encodedContent,
		Passphrase: keyPassphrase,
	}

	return &pvt, nil
}
