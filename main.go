package main

import (
	"context"
	"os"

	"github.com/smsimone/sshm/commands"

	"github.com/charmbracelet/fang"
)

func main() {
	if err := fang.Execute(context.Background(), commands.GetRootCommand()); err != nil {
		os.Exit(1)
	}
}
