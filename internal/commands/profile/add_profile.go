package profile

import (
	"encoding/base64"
	"fmt"
	"os"
	sshconn "sshm/internal/ssh_conn"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func AddProfileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-profile",
		Short: "Creates a new profile which can be used in a connection",
		RunE: func(cmd *cobra.Command, args []string) error {

			m := initialModel()
			res, err := tea.NewProgram(m).Run()
			if err != nil {
				return err
			} else if m.quitting {
				return fmt.Errorf("received null model from tea program")
			}

			m = res.(*model)

			profile, pvtKey := m.GetProfile()

			if pvtKey != nil && len(pvtKey.passphrase) > 0 {
				pvtKey, err := loadPrivateKey(pvtKey.key, &pvtKey.passphrase)
				if err != nil {
					return err
				}
				profile.PvtKey = pvtKey
			}

			return sshconn.AddProfile(profile)
		},
	}

	return cmd
}

func loadPrivateKey(privateKey string, keyPassphrase *string) (*sshconn.PrivateKey, error) {
	var content []byte

	if stat, err := os.Stat(privateKey); err != nil {
		fmt.Println("Provided key is not a file, trying to read as the content")
		content = []byte(privateKey)
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
			return nil, fmt.Errorf("failed to prase privatekey with password: %s", err.Error())
		}

		encodedContent := base64.StdEncoding.EncodeToString(content)
		pvt.Content = encodedContent
	} else {
		if _, err := ssh.ParsePrivateKey(content); err != nil {
			return nil, fmt.Errorf("failed to prase privatekey: %s", err.Error())
		}

		encodedContent := base64.StdEncoding.EncodeToString(content)
		pvt.Content = encodedContent
	}

	return &pvt, nil
}
