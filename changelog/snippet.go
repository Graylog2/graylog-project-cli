package changelog

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/pelletier/go-toml/v2"
	"github.com/samber/lo"
	"golang.org/x/exp/slices"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const TypeAdded = "added"
const TypeChanged = "changed"
const TypeDeprecated = "deprecated"
const TypeRemoved = "removed"
const TypeFixed = "fixed"
const TypeSecurity = "security"

var typePrefixMap = map[string]string{
	"a": TypeAdded,
	"c": TypeChanged,
	"d": TypeDeprecated,
	"r": TypeRemoved,
	"f": TypeFixed,
	"s": TypeSecurity,
}

// The available snippet types. The integer value represents the sort order.
var availableTypesMap = map[string]int{
	TypeAdded:      1,
	TypeChanged:    2,
	TypeDeprecated: 3,
	TypeRemoved:    4,
	TypeFixed:      5,
	TypeSecurity:   6,
}

var sortedTypes = lo.Keys(availableTypesMap)

func init() {
	slices.SortFunc(sortedTypes, func(a string, b string) int {
		if availableTypesMap[a] < availableTypesMap[b] {
			return -1
		} else if availableTypesMap[a] > availableTypesMap[b] {
			return 1
		}
		return 0
	})
}

type SnippetDetails struct {
	User      string `toml:"user"`
	Operators string `toml:"ops"`
}

type Snippet struct {
	Type          string         `toml:"type"`
	Message       string         `toml:"message"`
	Issues        []string       `toml:"issues"`
	PullRequests  []string       `toml:"pulls"`
	Contributors  []string       `toml:"contributors"`
	Details       SnippetDetails `toml:"details"`
	GitHubRepoURL string
	Filename      string
}

func listSnippets(path string) ([]string, error) {
	snippetFiles := make([]string, 0)

	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".toml") {
			return nil
		}

		snippetFiles = append(snippetFiles, path)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("couldn't traverse path %s: %w", path, err)
	}

	return snippetFiles, nil
}

func parseSnippet(snippetFile string, githubURL string) (*Snippet, error) {
	snippetBytes, err := os.ReadFile(snippetFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't read %s: %w", snippetFile, err)
	}

	if strings.TrimSpace(githubURL) == "" {
		githubURL, err = getGitHubURL(filepath.Dir(snippetFile))
		if err != nil {
			return nil, err
		}
	}

	var snippetData Snippet
	if err := toml.Unmarshal(snippetBytes, &snippetData); err != nil {
		return nil, fmt.Errorf("couldn't parse snippet %s: %w", snippetFile, err)
	}

	snippetData.GitHubRepoURL = githubURL
	snippetData.Filename = snippetFile

	for prefix, value := range typePrefixMap {
		if strings.HasPrefix(strings.ToLower(snippetData.Type), prefix) {
			snippetData.Type = value
		}
	}

	_, ok := availableTypesMap[snippetData.Type]
	if !ok {
		return nil, fmt.Errorf(`invalid type "%s" in file: %s`, snippetData.Type, snippetFile)
	}

	return &snippetData, nil
}

// Calls the Git binary to get the current repository URL.
func getGitHubURL(path string) (string, error) {
	urlString, err := git.GetRemoteUrl(path, "origin")
	if err != nil {
		return "", err
	}

	githubURL, err := utils.ParseGitHubURL(urlString)
	if err != nil {
		return "", fmt.Errorf("couldn't get Git repository URL: %w", err)
	}

	return githubURL.BrowserURL(), nil
}
