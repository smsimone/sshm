package commands

import (
	"github.com/smsimone/sshm/internal/commands/add"
	"github.com/smsimone/sshm/internal/commands/connect"
	"github.com/smsimone/sshm/internal/commands/list"
	"github.com/smsimone/sshm/internal/commands/profile"
	history "github.com/smsimone/sshm/internal/commands/versioning"

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
	rootCmd.AddCommand(history.NewCommand())
	rootCmd.AddCommand(profile.ListProfileCommand())
	return rootCmd
}
