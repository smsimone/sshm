package appstate

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	sshconn "term_cli/internal/ssh_conn"
)

const (
	configPath = "$HOME/.config/ssh_manager/config"
)

func ensureFile() {
	expanded := os.ExpandEnv(configPath)

	if _, err := os.Stat(expanded); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(expanded), 0755); err != nil {
			panic(fmt.Sprintf("Failed to create directory path: %s", err.Error()))
		}

		if _, err := os.Create(expanded); err != nil {
			panic(fmt.Sprintf("Failed to create config file: %s", err.Error()))
		}

		if err := os.WriteFile(expanded, []byte("[]"), os.ModeDevice); err != nil {
			panic(fmt.Sprintf("Failed to write basic data into file: %s", err.Error()))
		}
	}
}

func LoadFile() (*[]sshconn.Connection, error) {
	ensureFile()

	content, err := os.ReadFile(os.ExpandEnv(configPath))
	if err != nil {
		return nil, err
	}

	var data []sshconn.Connection
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return &data, nil
}

func AddItem(con sshconn.Connection) error {
	curr, err := LoadFile()
	if err != nil {
		return fmt.Errorf("failed to recover current items: %w", err)
	}
	data := *curr
	data = append(data, con)

	bytes, err := json.Marshal(&data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return os.WriteFile(os.ExpandEnv(configPath), bytes, os.ModeDevice)
}
