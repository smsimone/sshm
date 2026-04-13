package versioning

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/smsimone/sshm/internal/config"

	"github.com/go-git/go-git/v6"
	gitconfig "github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing/transport"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/go-git/go-git/v6/plumbing/transport/ssh"
	"github.com/go-git/go-git/v6/x/plugin"
	xconfig "github.com/go-git/go-git/v6/x/plugin/config"
)

var loadGitConfigOnce sync.Once

type CloneOptions struct {
	RepositoryPath string
	Url            string
	Force          bool
}

func loadGitConfig() {
	loadGitConfigOnce.Do(func() {
		_ = plugin.Register(plugin.ConfigLoader(), func() plugin.ConfigSource {
			globalConfig, err := gitconfig.LoadConfig(gitconfig.GlobalScope)
			if err != nil {
				globalConfig = gitconfig.NewConfig()
			}

			systemConfig, err := gitconfig.LoadConfig(gitconfig.SystemScope)
			if err != nil {
				systemConfig = gitconfig.NewConfig()
			}

			return xconfig.NewStatic(*globalConfig, *systemConfig)
		})
	})
}

func CloneRepository(ctx context.Context, opts CloneOptions) error {
	loadGitConfig()
	if opts.Force {
		if err := deleteDir(config.ConfigurationFolder()); err != nil {
			return fmt.Errorf("failed to delete configuration folder: %w", err)
		}
	}

	cloneOpts := &git.CloneOptions{
		URL:            opts.Url,
		Progress:       os.Stdout,
		Bare:           false,
		NoCheckout:     false,
		AllowEmptyRepo: true,
	}

	if auth, ok, err := authForRemote(opts.Url); err != nil {
		return err
	} else if ok {
		cloneOpts.Auth = auth
	}

	_, err := git.PlainCloneContext(ctx, opts.RepositoryPath, cloneOpts)
	return err
}

func PullRepository(repo *git.Repository) error {
	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	if err := wt.Pull(&git.PullOptions{}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("failed to update config repository: %w", err)
	}
	return nil
}

func PushRepository(repo *git.Repository) error {
	pushOpts, err := pushOptions(repo)
	if err != nil {
		return err
	}

	return repo.Push(pushOpts)
}

func pushOptions(repo *git.Repository) (*git.PushOptions, error) {
	remoteURL, ok, err := defaultRemoteURL(repo)
	if err != nil {
		return nil, err
	}

	pushOpts := &git.PushOptions{}
	if !ok {
		return pushOpts, nil
	}
	pushOpts.RemoteURL = remoteURL

	auth, hasAuth, err := authForRemote(remoteURL)
	if err != nil {
		return nil, err
	}
	if hasAuth {
		pushOpts.Auth = auth
	}

	return pushOpts, nil
}

func defaultRemoteURL(repo *git.Repository) (string, bool, error) {
	cfg, err := repo.Config()
	if err != nil {
		return "", false, fmt.Errorf("failed to read repository config: %w", err)
	}

	remote := cfg.Remotes[git.DefaultRemoteName]
	if remote == nil || len(remote.URLs) == 0 {
		return "", false, nil
	}

	return remote.URLs[0], true, nil
}

func authForRemote(rawURL string) (transport.AuthMethod, bool, error) {
	if isHTTPRemote(rawURL) {
		return authFromGitCredentialHelper(rawURL)
	}

	if isSSHRemote(rawURL) {
		auth, err := ssh.NewSSHAgentAuth(ssh.DefaultUsername)
		if err != nil {
			return nil, false, fmt.Errorf("failed to create ssh-agent auth: %w", err)
		}
		return auth, true, nil
	}

	return nil, false, nil
}

func isHTTPRemote(rawURL string) bool {
	if strings.Contains(rawURL, "://") {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return false
		}
		return parsedURL.Scheme == "http" || parsedURL.Scheme == "https"
	}

	return false
}

func isSSHRemote(rawURL string) bool {
	if strings.Contains(rawURL, "://") {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return false
		}

		switch parsedURL.Scheme {
		case "ssh", "git+ssh":
			return true
		default:
			return false
		}
	}

	if strings.Contains(rawURL, "@") && strings.Contains(rawURL, ":") {
		return true
	}

	return false
}

func authFromGitCredentialHelper(rawURL string) (transport.AuthMethod, bool, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse remote url: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, false, nil
	}

	request := fmt.Sprintf("protocol=%s\nhost=%s\n", parsedURL.Scheme, parsedURL.Host)
	path := strings.TrimPrefix(parsedURL.Path, "/")
	if path != "" {
		request += fmt.Sprintf("path=%s\n", path)
	}
	if parsedURL.User != nil {
		username := parsedURL.User.Username()
		if username != "" {
			request += fmt.Sprintf("username=%s\n", username)
		}
	}
	request += "\n"

	cmd := exec.Command("git", "credential", "fill")
	cmd.Stdin = strings.NewReader(request)
	output, err := cmd.Output()
	if err != nil {
		return nil, false, fmt.Errorf("failed to read credentials from git helper: %w", err)
	}

	values := map[string]string{}
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		values[parts[0]] = parts[1]
	}

	password := values["password"]
	if password == "" {
		return nil, false, nil
	}

	username := values["username"]
	if username == "" {
		username = "git"
	}

	return &githttp.BasicAuth{
		Username: username,
		Password: password,
	}, true, nil
}

func LoadRepository() *git.Repository {
	loadGitConfig()
	repo, _ := git.PlainOpen(config.ConfigurationFolder())
	return repo
}

func deleteDir(dirPath string) error {
	if stat, err := os.Stat(dirPath); err == nil && stat.IsDir() {
		return os.RemoveAll(dirPath)
	}
	return nil
}
