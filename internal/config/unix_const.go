//go:build !windows
// +build !windows

package config

import (
	"os"
	"path"
)

func configurationFilePath() string {
	return os.ExpandEnv(path.Join(configComponents...))
}
