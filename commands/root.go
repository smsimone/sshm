package commands

import (
	"term_cli/internal/commands/add"
	"term_cli/internal/commands/connect"
	"term_cli/internal/commands/list"
	"term_cli/internal/commands/profile"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "term",
	Short: "A simple example program!",
}

func GetRootCommand() *cobra.Command {
	rootCmd.AddCommand(add.NewTeaCommand())
	rootCmd.AddCommand(list.NewCommand())
	rootCmd.AddCommand(connect.NewCommand())
	rootCmd.AddCommand(profile.AddProfileCommand())
	return rootCmd
}
