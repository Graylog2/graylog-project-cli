package utils

import (
	"fmt"
	"github.com/pkg/errors"
	neturl "net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func ParseGitHubURL(url string) (GitHubURL, error) {
	if !strings.HasSuffix(url, ".git") {
		return GitHubURL{}, errors.Errorf("GitHub URL is missing .git suffix: %s", url)
	}

	var repository string
	switch {
	case strings.HasPrefix(url, "github://"):
		repository = strings.Split(url, "//")[1]
	case strings.HasPrefix(url, "git@github"):
		repository = strings.Split(url, ":")[1]
	case strings.HasPrefix(url, "https://"):
		repository = strings.Split(url, "github.com/")[1]
	default:
		return GitHubURL{}, errors.Errorf("unknown GitHub URL: %s", url)
	}

	return CreateGitHubURL(url, repository)
}

func ParseGitHubPRString(prString string) (string, int, error) {
	pattern := regexp.MustCompile("^[\\w-_.]+/[\\w-_.]+#\\d+$")

	var parts []string

	if strings.HasPrefix(strings.ToLower(prString), "https://api.github.com/") && strings.Contains(prString, "/pulls/") {
		// Input is a PR API URL like this: https://api.github.com/repos/Graylog2/graylog2-server/pulls/9309
		u, err := neturl.Parse(prString)
		if err != nil {
			return "", 0, errors.Wrapf(err, "couldn't parse GitHub pull request API URL <%s>", prString)
		}
		parts = strings.SplitN(strings.TrimPrefix(u.Path, "/repos/"), "/pulls/", 2)
	} else if strings.HasPrefix(strings.ToLower(prString), "https://github.com/") && strings.Contains(prString, "/pull/") {
		// Input is a PR URL like this: https://github.com/Graylog2/graylog2-server/pull/9692
		u, err := neturl.Parse(prString)
		if err != nil {
			return "", 0, errors.Wrapf(err, "couldn't parse GitHub pull request URL <%s>", prString)
		}
		parts = strings.SplitN(strings.TrimPrefix(u.Path, "/"), "/pull/", 2)
	} else if pattern.MatchString(prString) {
		// Input is "<owner>/<repo>#<pr-num>" (e.g. Graylog2/graylog2-server#123)
		parts = strings.SplitN(prString, "#", 2)
	} else {
		return "", 0, errors.Errorf("unknown GitHub pull request string <%s>", prString)
	}

	if len(parts) != 2 {
		return "", 0, errors.Errorf("couldn't extract pull request repo and number from <%s> (%v)", prString, parts)
	}

	prRepo := parts[0]
	if prRepo == "" {
		return "", 0, errors.Errorf("couldn't parse pull request repository from <%s>", prString)
	}
	prNumber, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, errors.Errorf("couldn't parse pull request number from <%s>", prString)
	}

	return prRepo, prNumber, nil
}

func ResolveGitHubIssueURL(baseRepo string, issueString string) (string, error) {
	repo := strings.TrimSuffix(baseRepo, ".git")
	fullURLPattern := regexp.MustCompile("^https://github\\.com/[^/]+/[^/]+/(?:issues|pull)/\\d+$")
	numPattern := regexp.MustCompile("^#?(\\d+)$")
	repoNamePattern := regexp.MustCompile("^([^/\\s]+)#(\\d+)$")
	orgRepoNamePattern := regexp.MustCompile("^([^/\\s]+)/([^/\\s]+)#(\\d+)$")

	if fullURLPattern.FindStringSubmatch(issueString) != nil {
		return issueString, nil
	}
	if matches := numPattern.FindStringSubmatch(issueString); matches != nil {
		return repo + "/issues/" + matches[1], nil
	}
	if matches := repoNamePattern.FindStringSubmatch(issueString); matches != nil {
		org := strings.Split(strings.TrimPrefix(repo, "https://github.com/"), "/")[0]
		return "https://github.com/" + org + "/" + matches[1] + "/issues/" + matches[2], nil
	}
	if matches := orgRepoNamePattern.FindStringSubmatch(issueString); matches != nil {
		return "https://github.com/" + matches[1] + "/" + matches[2] + "/issues/" + matches[3], nil
	}

	return "", fmt.Errorf("couldn't parse issue string \"%s\" for repository \"%s\"", issueString, baseRepo)
}

type PrettyMode int

const (
	PrettyModeNum PrettyMode = iota
	PrettyModeRepo
	PrettyModeOrgRepo
)

func PrettifyGitHubIssueURL(githubURL string, mode PrettyMode) string {
	pattern := regexp.MustCompile("^https://github\\.com/([^/]+)/([^/]+)/(?:issues|pull)/(\\d+)$")
	match := pattern.FindStringSubmatch(githubURL)

	switch mode {
	case PrettyModeNum:
		return fmt.Sprintf("#%s", match[3])
	case PrettyModeRepo:
		return fmt.Sprintf("%s#%s", match[2], match[3])
	case PrettyModeOrgRepo:
		fallthrough
	default:
		return fmt.Sprintf("%s/%s#%s", match[1], match[2], match[3])
	}
}

func ReplaceGitHubURL(url string, repoName string) (string, error) {
	name := strings.TrimSuffix(repoName, ".git")

	switch {
	case strings.HasPrefix(url, "github://"):
		return fmt.Sprintf("github://%s.git", name), nil
	case strings.HasPrefix(url, "git@github"):
		return fmt.Sprintf("git@github.com:%s.git", name), nil
	case strings.HasPrefix(url, "https://github"):
		return fmt.Sprintf("https://github.com/%s.git", name), nil
	default:
		return "", errors.Errorf("unknown GitHub URL: %s", url)
	}
}

func CreateGitHubURL(url string, repository string) (GitHubURL, error) {
	if url == "" || repository == "" {
		return GitHubURL{}, fmt.Errorf("url and repository cannot be empty")
	}
	if strings.HasPrefix(url, "https://") {
		u, err := neturl.Parse(url)
		if err != nil {
			return GitHubURL{}, fmt.Errorf("couldn't parse URL: %s", err)
		}
		u.User = nil
		return GitHubURL{url: u.Redacted(), repository: repository}, nil
	}
	return GitHubURL{url: url, repository: repository}, nil
}

type GitHubURL struct {
	url        string
	repository string
}

func (url GitHubURL) URL() string {
	return url.url
}

func (url GitHubURL) Repository() string {
	return url.repository
}

func (url GitHubURL) IsHTTPS() bool {
	return strings.HasPrefix(url.url, "https://")
}

func (url GitHubURL) IsSSH() bool {
	return strings.HasPrefix(url.url, "git@")
}

func (url GitHubURL) SSH() string {
	return "git@github.com:" + url.repository
}

func (url GitHubURL) HTTPS() string {
	return "https://github.com/" + url.repository
}

func (url GitHubURL) Directory() string {
	return strings.TrimSuffix(filepath.Base(url.repository), filepath.Ext(url.repository))
}

func (url GitHubURL) BrowserURL() string {
	return strings.TrimSuffix(url.HTTPS(), filepath.Ext(url.repository))
}

func (url GitHubURL) Matches(match string) bool {
	repoName := strings.TrimSuffix(url.repository, filepath.Ext(url.repository))
	return strings.Compare(strings.ToLower(repoName), strings.ToLower(match)) == 0
}

func (url GitHubURL) String() string {
	return "github://" + url.repository
}
