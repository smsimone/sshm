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

var state appState = appState{}

type appState struct {
	Connections []Connection `json:"connections"`
	Profiles    []Profile    `json:"profiles"`
}

func (as *appState) persist() error {
	bytes, err := json.Marshal(*as)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return os.WriteFile(os.ExpandEnv(configPath), bytes, os.ModeDevice)
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

func loadFile() error {
	ensureFile()

	content, err := os.ReadFile(os.ExpandEnv(configPath))
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &state); err != nil {
		return fmt.Errorf("failed to unmarshal file: %w", err)
	}

	return nil
}

func GetProfile(label string) (*Profile, error) {
	err := loadFile()
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

func LoadProfiles() (*[]Profile, error) {
	err := loadFile()
	if err != nil {
		return nil, err
	}
	return &state.Profiles, nil
}

func LoadConnections() (*[]Connection, error) {
	err := loadFile()
	if err != nil {
		return nil, err
	}
	return &state.Connections, nil
}

func AddItem(con Connection) error {
	err := loadFile()
	if err != nil {
		return fmt.Errorf("failed to recover current items: %w", err)
	}
	state.Connections = append(state.Connections, con)
	return state.persist()
}

func AddProfile(profile Profile) error {
	err := loadFile()
	if err != nil {
		return fmt.Errorf("failed to recover current items: %w", err)
	}

	state.Profiles = append(state.Profiles, profile)

	return state.persist()
}
