package cli_updater

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strings"
)

var (
	CliVersion = "v0.0.1"
)

const (
	cliRepository = "http://100.107.123.81:3000/api/v1/repos/iamsmaso/sshm_cli/releases/latest"
)

type Release struct {
	TagName string   `json:"tag_name"`
	HTMLURL string   `json:"html_url"`
	Assets  *[]Asset `json:"assets"`
}

type Asset struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	DownloadUrl string `json:"browser_download_url"`
}

func (a *Asset) CanDownload() bool {
	osName := runtime.GOOS
	return strings.Contains(a.Name, strings.ReplaceAll(osName, "/", "-"))
}

func GetLatestVersion() *Release {
	resp, err := http.Get(cliRepository)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil
	}

	return &release
}
