package changelog

import (
	"bytes"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/pelletier/go-toml/v2"
	"io/fs"
	"io/ioutil"
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
	Type         string         `toml:"type"`
	Message      string         `toml:"message"`
	Issues       []string       `toml:"issues"`
	PullRequests []string       `toml:"pulls"`
	Contributors []string       `toml:"contributors"`
	Details      SnippetDetails `toml:"details"`
}

func Render(format string, path string) error {
	parsedSnippets, err := parseSnippets(path)
	if err != nil {
		return err
	}

	renderer, err := GetRenderer(format)
	if err != nil {
		return err
	}

	for _, _type := range typeOrder {
		if len(parsedSnippets[_type]) > 0 {
			buf := bytes.Buffer{}

			if err := renderer.RenderType(_type, &buf); err != nil {
				return fmt.Errorf("couldn't render type \"%s\": %w", _type, err)
			}

			if err := renderer.RenderSnippets(parsedSnippets[_type], &buf); err != nil {
				return fmt.Errorf("couldn't render snippets \"%#v\": %w", parsedSnippets[_type], err)
			}

			fmt.Println(buf.String())
		}
	}

	return nil
}

func parseSnippets(path string) (map[string][]Snippet, error) {
	snippetFiles := make([]string, 0)
	parsedSnippets := make(map[string][]Snippet)

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

	for _, snippetFile := range snippetFiles {
		snippetBytes, err := ioutil.ReadFile(snippetFile)
		if err != nil {
			return parsedSnippets, fmt.Errorf("couldn't read %s: %w", snippetFile, err)
		}

		var snippetData Snippet
		if err := toml.Unmarshal(snippetBytes, &snippetData); err != nil {
			return parsedSnippets, fmt.Errorf("couldn't parse snippet %s: %w", snippetFile, err)
		}

		for prefix, value := range typePrefixMap {
			if strings.HasPrefix(strings.ToLower(snippetData.Type), prefix) {
				snippetData.Type = value
			}
		}

		parsedSnippets[snippetData.Type] = append(parsedSnippets[snippetData.Type], snippetData)
	}

	return parsedSnippets, nil
}
