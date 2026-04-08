package profile

import (
	"encoding/base64"
	"fmt"
	"os"
	sshconn "term_cli/internal/ssh_conn"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func AddProfileCommand() *cobra.Command {
	var (
		label         string
		username      string
		password      string
		privateKey    string
		keyPassphrase string
	)

	cmd := &cobra.Command{
		Use:   "add-profile",
		Short: "Creates a new profile which can be used in a connection",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := sshconn.GetProfile(label); err != nil {
				return nil
			}
			return fmt.Errorf("Label '%s' already defined", label)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := sshconn.Profile{Label: label, Username: username}

			if len(privateKey) > 0 {
				pvtKey, err := loadPrivateKey(privateKey, &keyPassphrase)
				if err != nil {
					return err
				}
				profile.PvtKey = pvtKey
			} else if len(password) > 0 {
				profile.Password = &password
			}

			return sshconn.AddProfile(profile)
		},
	}

	cmd.Flags().StringVarP(&label, "label", "l", "", "Label to identify the profile")
	cmd.Flags().StringVarP(&username, "username", "u", os.Getenv("USER"), "Username which must be used to login")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	cmd.Flags().StringVarP(&privateKey, "private_key", "P", "", "Private key (file/content)")
	cmd.Flags().StringVarP(&keyPassphrase, "passphrase", "k", "", "Passophrase to decrypt the ssh key")

	markRequired(cmd, "label", "username")

	return cmd
}

func markRequired(cmd *cobra.Command, label ...string) {
	for _, labl := range label {
		if err := cmd.MarkFlagRequired(labl); err != nil {
			panic(err)
		}
	}
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
