package config

import (
	"path"
	"sync"
)

var (
	confPathOnce     sync.Once
	configComponents = []string{"$HOME", ".config", "ssh_manager", "config"}
	confPath         string
)

func ConfigurationFilePath() string {
	confPathOnce.Do(func() {
		confPath = configurationFilePath()
	})
	return confPath
}

func ConfigurationFolder() string {
	return path.Dir(ConfigurationFilePath())
}
