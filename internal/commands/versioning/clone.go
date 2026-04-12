package history

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/smsimone/sshm/internal/config"
	"github.com/smsimone/sshm/internal/versioning"

	"github.com/go-git/go-git/v6"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var (
		force bool
	)

	cmd := &cobra.Command{
		Use:  "clone <repository>",
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			_, err := url.Parse(args[0])
			if err != nil {
				return fmt.Errorf("Invalid url: %s", err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			err := versioning.CloneRepository(ctx, versioning.CloneOptions{
				RepositoryPath: config.ConfigurationFolder(),
				Url:            args[0],
				Force:          force,
			})
			if errors.Is(err, git.ErrTargetDirNotEmpty) {
				return fmt.Errorf("Directory not empty. Should use --force instead")
			} else {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Forces the clone")

	return cmd
}
