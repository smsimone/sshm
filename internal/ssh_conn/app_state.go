package sshconn

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
)

const (
	configPath = "$HOME/.config/ssh_manager/config"
)

type appState struct {
	Connections []Connection `json:"connections"`
	Profiles    []Profile    `json:"profiles"`
}

func ensureFile() {
	expanded := os.ExpandEnv(configPath)

	if _, err := os.Stat(expanded); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(expanded), 0755); err != nil {
			panic(fmt.Sprintf("Failed to create directory path: %s", err.Error()))
		}

		if _, err := os.Create(expanded); err != nil {
			panic(fmt.Sprintf("Failed to create config file: %s", err.Error()))
		}

		if err := os.WriteFile(expanded, []byte("{}"), os.ModeDevice); err != nil {
			panic(fmt.Sprintf("Failed to write basic data into file: %s", err.Error()))
		}
	}
}

func LoadFile() (*appState, error) {
	ensureFile()

	content, err := os.ReadFile(os.ExpandEnv(configPath))
	if err != nil {
		return nil, err
	}

	var data appState
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return &data, nil
}

func GetProfile(label string) (*Profile, error) {
	state, err := LoadFile()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %s", err.Error())
	}

	idx := slices.IndexFunc(state.Profiles, func(p Profile) bool {
		return p.Label == label
	})
	if idx == -1 {
		return nil, fmt.Errorf("profile %s not found", label)
	}

	return &state.Profiles[idx], nil
}

func LoadConnections() (*[]Connection, error) {
	state, err := LoadFile()
	if err != nil {
		return nil, err
	}
	return &state.Connections, nil
}

func AddItem(con Connection) error {
	curr, err := LoadFile()
	if err != nil {
		return fmt.Errorf("failed to recover current items: %w", err)
	}
	data := *curr
	data.Connections = append(data.Connections, con)

	bytes, err := json.Marshal(&data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return os.WriteFile(os.ExpandEnv(configPath), bytes, os.ModeDevice)
}
