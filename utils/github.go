package utils

import (
	"github.com/pkg/errors"
	"path/filepath"
	"strings"
)

func ParseGitHubURL(url string) (GitHubURL, error) {
	gitHubURL := GitHubURL{}

	if !strings.HasSuffix(url, ".git") {
		return gitHubURL, errors.Errorf("GitHub URL is missing .git suffix: %s", url)
	}

	switch {
	case strings.HasPrefix(url, "github://"):
		gitHubURL.Repository = strings.Split(url, "//")[1]
	case strings.HasPrefix(url, "git@github"):
		gitHubURL.Repository = strings.Split(url, ":")[1]
	case strings.HasPrefix(url, "https://github"):
		gitHubURL.Repository = strings.Split(url, "github.com/")[1]
	default:
		return GitHubURL{}, errors.Errorf("unknown GitHub URL: %s", url)
	}

	return gitHubURL, nil
}

type GitHubURL struct {
	Repository string
}

func (url GitHubURL) SSH() string {
	return "git@github.com:" + url.Repository
}

func (url GitHubURL) HTTPS() string {
	return "https://github.com/" + url.Repository
}

func (url GitHubURL) Directory() string {
	return strings.TrimSuffix(filepath.Base(url.Repository), filepath.Ext(url.Repository))
}

func (url GitHubURL) String() string {
	return "github://" + url.Repository
}
