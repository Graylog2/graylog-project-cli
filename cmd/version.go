package cmd

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Graylog2/graylog-project-cli/config"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/fatih/color"
	"github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
)

type GithubRelease struct {
	TagName     string `json:"tag_name"`
	HtmlUrl     string `json:"html_url"`
	PublishedAt string `json:"published_at"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version",
	Long:  "Display version and check for updates if not disabled",
	Run:   versionCommand,
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

func versionCommand(cmd *cobra.Command, args []string) {
	config := config.Get()
	logger.Info("Version %v, built on %v (revision %v)", gitTag, buildDate, gitRevision)
	logger.Info("Test test test!!")

	if !config.NoUpdateCheck {
		printState(getLatestRelease(), true)
	}
}

func printState(latestRelease GithubRelease, onlyOutdated bool) {
	if latestRelease.TagName == "" {
		return
	}
	ourVersion, _ := version.NewVersion(gitTag)
	publicVersion, _ := version.NewVersion(latestRelease.TagName)
	if ourVersion.LessThan(publicVersion) {
		logger.ColorInfo(color.FgRed, "\nYou are running an outdated version!\n")
		logger.Info("Current release version is: %v (released on %v) available at %v", latestRelease.TagName, latestRelease.PublishedAt, latestRelease.HtmlUrl)
	} else if !onlyOutdated {
		logger.ColorInfo(color.FgGreen, "You are running the latest version.")
	}
}

func getLatestRelease() GithubRelease {
	resp, err := http.Get("https://api.github.com/repos/graylog2/graylog-project-cli/releases/latest")
	if err != nil || resp.StatusCode >= 400 {
		return GithubRelease{}
	}

	var latestRelease GithubRelease
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &latestRelease)
	return latestRelease
}

func CheckForUpdate() {
	config := config.Get()
	if !config.NoUpdateCheck {
		printState(getLatestRelease(), true)
	}
}
