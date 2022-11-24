package changelog

import (
	"bytes"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/pelletier/go-toml/v2"
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

var typeOrder = []string{TypeAdded, TypeChanged, TypeDeprecated, TypeRemoved, TypeFixed, TypeSecurity}

var typePrefixMap = map[string]string{
	"a": TypeAdded,
	"c": TypeChanged,
	"d": TypeDeprecated,
	"r": TypeRemoved,
	"f": TypeFixed,
	"s": TypeSecurity,
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

func Render(config Config) error {
	parsedSnippets, err := parseSnippets(config.SnippetsPaths)
	if err != nil {
		return err
	}

	renderer, err := GetRenderer(config.RenderFormat)
	if err != nil {
		return err
	}

	headBuf := bytes.Buffer{}
	if err := renderer.RenderHeader(config, &headBuf); err != nil {
		return fmt.Errorf("couldn't render header: %w", err)
	}
	fmt.Print(headBuf.String())

	for _, _type := range typeOrder {
		if len(parsedSnippets[_type]) > 0 {
			buf := bytes.Buffer{}

			if err := renderer.RenderType(config, _type, &buf); err != nil {
				return fmt.Errorf("couldn't render type \"%s\": %w", _type, err)
			}

			if err := renderer.RenderSnippets(config, parsedSnippets[_type], &buf); err != nil {
				return fmt.Errorf("couldn't render snippets: %w", err)
			}

			fmt.Println(buf.String())
		}
	}

	return nil
}

func getGitHubURL(path string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("couldn't get Git repository URL: %w", err)
	}
	defer os.Chdir(cwd)

	if err := os.Chdir(path); err != nil {
		return "", fmt.Errorf("couldn't get Git repository URL: %w", err)
	}

	urlString, err := git.GitValueE("remote", "get-url", "--push", "origin")
	if err != nil {
		return "", fmt.Errorf("couldn't get Git repository URL: %w", err)
	}

	githubURL, err := utils.ParseGitHubURL(urlString)
	if err != nil {
		return "", fmt.Errorf("couldn't get Git repository URL: %w", err)
	}

	return githubURL.BrowserURL(), nil
}

func parseSnippets(paths []string) (map[string][]Snippet, error) {
	parsedSnippets := make(map[string][]Snippet)

	for _, path := range paths {
		snippetFiles := make([]string, 0)

		err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".toml") {
				return nil
			}

			logger.Debug("Adding snippetFile: %s", path)
			snippetFiles = append(snippetFiles, path)

			return nil
		})

		if err != nil {
			return parsedSnippets, fmt.Errorf("couldn't traverse path %s: %w", path, err)
		}

		githubURL, err := getGitHubURL(path)
		if err != nil {
			return parsedSnippets, err
		}

		for _, snippetFile := range snippetFiles {
			snippetBytes, err := os.ReadFile(snippetFile)
			if err != nil {
				return parsedSnippets, fmt.Errorf("couldn't read %s: %w", snippetFile, err)
			}

			var snippetData Snippet
			if err := toml.Unmarshal(snippetBytes, &snippetData); err != nil {
				return parsedSnippets, fmt.Errorf("couldn't parse snippet %s: %w", snippetFile, err)
			}

			snippetData.GitHubRepoURL = githubURL
			snippetData.Filename = snippetFile

			for prefix, value := range typePrefixMap {
				if strings.HasPrefix(strings.ToLower(snippetData.Type), prefix) {
					snippetData.Type = value
				}
			}

			parsedSnippets[snippetData.Type] = append(parsedSnippets[snippetData.Type], snippetData)
		}
	}

	return parsedSnippets, nil
}
