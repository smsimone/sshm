package main

import (
	"context"
	"os"

	"github.com/smsimone/sshm/commands"

	"github.com/charmbracelet/fang"
)

const (
	version string = "v0.0.4"
)

func main() {
	if err := fang.Execute(context.Background(), commands.GetRootCommand(), fang.WithVersion(version)); err != nil {
		os.Exit(1)
	}
}
