package updater

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"github.com/smsimone/sshm/internal/cli_updater"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var (
		dryRun bool
	)

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Updates the binary with the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			latest := cli_updater.GetLatestVersion()
			if latest.TagName == cli_updater.CliVersion {
				fmt.Println("You are already on the latest version")
				return nil
			}
			if latest.Assets != nil && len(*latest.Assets) == 0 {
				return fmt.Errorf("No assets available to download")
			}

			var downloadable *cli_updater.Asset
			for _, a := range *latest.Assets {
				if a.CanDownload() {
					if downloadable != nil {
						return fmt.Errorf("Too many downloadable assets. Install manually")
					}
					downloadable = &a
				}
			}
			if downloadable == nil {
				return fmt.Errorf("No downloadable asset found")
			}

			fmt.Println("Downloading asset")

			resp, err := http.Get(downloadable.DownloadUrl)
			if err != nil || resp.StatusCode != 200 {
				return fmt.Errorf("Failed to download asset")
			}
			defer resp.Body.Close()
			fmt.Println("Asset downloaded")

			data, _ := io.ReadAll(resp.Body)
			fmt.Println("Read all bytes from body")
			return os.WriteFile("./sshm_latest", data, os.ModeCharDevice)
		},
	}

	cmd.Flags().BoolVarP(&dryRun, "dry-run", "", false, "Checks only the presence of an update. Does not download anything")

	return cmd
}
