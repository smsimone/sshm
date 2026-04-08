package add

import (
	"encoding/base64"
	"fmt"
	"os"

	sshconn "term_cli/internal/ssh_conn"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func NewTeaCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add-connection",
		Short: "Add a new connection to the file",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := initialModel()
			_, err := tea.NewProgram(p).Run()
			if err != nil {
				return err
			}

			return nil
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

			//			if len(privateKey) > 0 {
			//				pvtKey, err := loadPrivateKey(privateKey, &keyPassphrase)
			//				if err != nil {
			//					return err
			//				}
			//				conn.PvtKey = pvtKey
			//			} else if len(password) > 0 {
			//				conn.Password = &password
			//			}

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

func loadPrivateKey(privateKey string, keyPassphrase *string) (*sshconn.PrivateKey, error) {
	var content []byte

	if stat, err := os.Stat(privateKey); err != nil {

		fmt.Println("Provided key is not a file, trying to read as the content")
		content = []byte(privateKey)

		return nil, fmt.Errorf("private key does not exists")
	} else if stat.IsDir() {
		return nil, fmt.Errorf("private key path is pointing to a directory")
	} else {
		if fileContent, err := os.ReadFile(privateKey); err != nil {
			return nil, err
		} else {
			content = fileContent
		}
	}

	pvt := sshconn.PrivateKey{}
	if keyPassphrase != nil && len(*keyPassphrase) > 0 {
		pvt.Passphrase = keyPassphrase
	}

	if pvt.Passphrase != nil {
		if _, err := ssh.ParsePrivateKeyWithPassphrase(content, []byte(*pvt.Passphrase)); err != nil {
			return nil, err
		}

		encodedContent := base64.StdEncoding.EncodeToString(content)
		pvt.Content = encodedContent
	} else {
		if _, err := ssh.ParsePrivateKey(content); err != nil {
			return nil, err
		}

		encodedContent := base64.StdEncoding.EncodeToString(content)
		pvt.Content = encodedContent
	}

	return &pvt, nil
}
