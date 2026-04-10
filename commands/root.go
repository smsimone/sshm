package commands

import (
	"sshm/internal/cli_updater"
	"sshm/internal/commands/add"
	"sshm/internal/commands/connect"
	"sshm/internal/commands/list"
	"sshm/internal/commands/profile"
	"sshm/internal/commands/updater"
	history "sshm/internal/commands/versioning"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "sshm",
	Short:   "Ssh connectoin manager",
	Version: cli_updater.CliVersion,
}

func GetRootCommand() *cobra.Command {
	rootCmd.AddCommand(add.NewCommand())
	rootCmd.AddCommand(list.NewCommand())
	rootCmd.AddCommand(connect.NewCommand())
	rootCmd.AddCommand(profile.AddProfileCommand())
	rootCmd.AddCommand(history.NewCommand())
	rootCmd.AddCommand(profile.ListProfileCommand())
	rootCmd.AddCommand(updater.NewCommand())
	return rootCmd
}
