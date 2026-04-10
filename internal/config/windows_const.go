//go:build windows
// +build windows

package config

import (
	"os"
	"path"
)

func configurationFilePath() string {
	return os.Expand(path.Join(configComponents...), func(component string) string {
		switch component {
		case "HOME":
			return os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		default:
			return os.Getenv(component)
		}
	})
}
