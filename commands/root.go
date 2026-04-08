package commands

import (
	"sshm/internal/commands/add"
	"sshm/internal/commands/connect"
	"sshm/internal/commands/list"
	"sshm/internal/commands/profile"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "sshm",
	Short:   "Ssh connectoin manager",
	Version: "0.0.1",
}

func GetRootCommand() *cobra.Command {
	rootCmd.AddCommand(add.NewTeaCommand())
	rootCmd.AddCommand(list.NewCommand())
	rootCmd.AddCommand(connect.NewCommand())
	rootCmd.AddCommand(profile.AddProfileCommand())
	return rootCmd
}
