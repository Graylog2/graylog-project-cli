package changelog

import (
	"bytes"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
)

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

	for _, _type := range sortedTypes {
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

func parseSnippets(paths []string) (map[string][]Snippet, error) {
	parsedSnippets := make(map[string][]Snippet)

	for _, path := range paths {
		snippetFiles, err := listSnippets(path)
		if err != nil {
			return parsedSnippets, err
		}

		for _, snippetFile := range snippetFiles {
			logger.Debug("Parsing file %s", snippetFile)

			snippetData, err := parseSnippet(snippetFile)
			if err != nil {
				return parsedSnippets, err
			}

			parsedSnippets[snippetData.Type] = append(parsedSnippets[snippetData.Type], *snippetData)
		}
	}

	return parsedSnippets, nil
}
