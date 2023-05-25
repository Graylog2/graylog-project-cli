package changelog

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/samber/lo"
	"os"
)

func Render(config Config) error {
	parsedSnippets, err := parseSnippets(config)
	if err != nil {
		return err
	}

	renderer, err := GetRenderer(config.RenderFormat)
	if err != nil {
		return err
	}

	if !config.SkipHeader {
		headBuf := bytes.Buffer{}
		if err := renderer.RenderHeader(config, &headBuf); err != nil {
			return fmt.Errorf("couldn't render header: %w", err)
		}
		fmt.Print(headBuf.String())
	}

	numSnippets := lo.Sum(lo.Map(lo.Values(parsedSnippets), func(item []Snippet, index int) int {
		return len(item)
	}))

	if numSnippets < 1 && config.RenderNoChanges {
		noChangeBuf := bytes.Buffer{}
		if err := renderer.RenderNoChanges(config, &noChangeBuf); err != nil {
			return fmt.Errorf("couldn't render no-changes paragraph: %w", err)
		}
		fmt.Print(noChangeBuf.String())
		return nil
	}

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

func parseSnippets(config Config) (map[string][]Snippet, error) {
	parsedSnippets := make(map[string][]Snippet)
	paths := config.SnippetsPaths
	stdin := config.ReadStdin

	if stdin {
		paths = make([]string, 0)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			paths = append(paths, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("couldn't read snippet files from STDIN: %w", err)
		}
	}

	for _, path := range paths {
		snippetFiles, err := listSnippets(path)
		if err != nil {
			return parsedSnippets, err
		}

		for _, snippetFile := range snippetFiles {
			logger.Debug("Parsing file %s", snippetFile)

			snippetData, err := parseSnippet(snippetFile)
			if err != nil {
				if config.SkipInvalidSnippets {
					logger.Info("Skipping invalid snippet file: %s", err)
					continue
				}
				return parsedSnippets, err
			}

			parsedSnippets[snippetData.Type] = append(parsedSnippets[snippetData.Type], *snippetData)
		}
	}

	return parsedSnippets, nil
}
