package config

import (
	"os"
	"path"
)

const (
	configPath = "$HOME/.config/ssh_manager/config"
)

func ConfigurationFilePath() string {
	return os.ExpandEnv(configPath)
}

func ConfigurationFolder() string {
	return path.Dir(ConfigurationFilePath())
}
