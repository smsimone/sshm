package main

import (
	"context"
	"os"
	"term_cli/commands"

	"github.com/charmbracelet/fang"
)

func main() {
	if err := fang.Execute(context.Background(), commands.GetRootCommand()); err != nil {
		os.Exit(1)
	}
}
