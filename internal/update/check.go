package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// Check compares the current version against the latest GitHub release.
// Returns the latest version string if newer, empty string otherwise.
func Check(currentVersion string) string {
	if os.Getenv("SPRITZ_NO_UPDATE_CHECK") != "" {
		return ""
	}
	if currentVersion == "dev" {
		return ""
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/spritz-finance/spritz-cli/releases/latest")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(currentVersion, "v")

	if latest != "" && latest != current {
		return latest
	}
	return ""
}

// PrintNotice writes an update notice to stderr if a newer version is available.
func PrintNotice(currentVersion, latestVersion string) {
	if latestVersion == "" {
		return
	}
	fmt.Fprintf(os.Stderr, "\nA new version of spritz is available: %s → %s\nRun 'spritz update' to upgrade.\n",
		currentVersion, latestVersion)
}
