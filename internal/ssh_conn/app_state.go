package sshconn

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/smsimone/sshm/internal/config"
	"github.com/smsimone/sshm/internal/versioning"

	"github.com/go-git/go-git/v6"
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

	err = os.WriteFile(config.ConfigurationFilePath(), bytes, os.ModeDevice)
	if err != nil {
		return err
	}

	go func() {
		commitFiles()
	}()

	return nil
}

func commitFiles() error {
	if repo := versioning.LoadRepository(); repo != nil {
		wt, err := repo.Worktree()
		if err != nil {
			return fmt.Errorf("failed to open worktree: %w", err)
		}
		if err := wt.AddGlob("."); err != nil {
			return fmt.Errorf("failed to add files: %w", err)
		}
		_, err = wt.Commit("Added configuration", &git.CommitOptions{})
		if err != nil {
			return fmt.Errorf("failed to create commit: %w", err)
		}
		return versioning.PushRepository(repo)
	}
	return nil
}

func ensureFile() {
	if _, err := os.Stat(config.ConfigurationFilePath()); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(config.ConfigurationFolder(), 0755); err != nil {
			panic(fmt.Sprintf("Failed to create directory path: %s", err.Error()))
		}

		if _, err := os.Create(config.ConfigurationFilePath()); err != nil {
			panic(fmt.Sprintf("Failed to create config file: %s", err.Error()))
		}

		if err := os.WriteFile(config.ConfigurationFilePath(), []byte("{}"), os.ModeDevice); err != nil {
			panic(fmt.Sprintf("Failed to write basic data into file: %s", err.Error()))
		}
	}
}

func loadFile() error {
	ensureFile()

	if repo, _ := git.PlainOpen(config.ConfigurationFolder()); repo != nil {
		if err := versioning.PullRepository(repo); err != nil {
			return fmt.Errorf("failed to pull updates: %w", err)
		}
	}

	content, err := os.ReadFile(config.ConfigurationFilePath())
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
	if idx := slices.IndexFunc(state.Connections, func(p Connection) bool { return p.Label == con.Label }); idx != -1 {
		return fmt.Errorf("Label '%s' is in use", con.Label)
	}

	state.Connections = append(state.Connections, con)
	return state.persist()
}

func AddProfile(profile Profile) error {
	err := loadFile()
	if err != nil {
		return fmt.Errorf("failed to recover current items: %w", err)
	}

	if idx := slices.IndexFunc(state.Profiles, func(p Profile) bool { return p.Label == profile.Label }); idx != -1 {
		return fmt.Errorf("Label '%s' is in use", profile.Label)
	}

	state.Profiles = append(state.Profiles, profile)

	return state.persist()
}

func RemoveProfile(label string, force bool) error {
	if err := loadFile(); err != nil {
		return fmt.Errorf("Failed to recover current items: %w", err)
	}

	profileIdx := slices.IndexFunc(state.Profiles, func(p Profile) bool { return p.Label == label })
	if profileIdx == -1 {
		return fmt.Errorf("Label '%s' does not exists", label)
	}
	profile := state.Profiles[profileIdx]

	connIdx := slices.IndexFunc(state.Connections, func(p Connection) bool { return p.Profile != nil && *p.Profile == profile.Label })
	if connIdx != -1 && !force {
		return fmt.Errorf("Profile in use. Must use force (will delete also the connection)")
	}

	// removes the connection
	if connIdx != -1 {
		tmp := []Connection{}
		for x, conn := range state.Connections {
			if x != connIdx {
				tmp = append(tmp, conn)
			}
		}
		state.Connections = tmp
	}

	tmp := []Profile{}
	for x, profile := range state.Profiles {
		if x != profileIdx {
			tmp = append(tmp, profile)
		}
	}
	state.Profiles = tmp

	return state.persist()
}
