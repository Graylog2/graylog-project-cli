package changelog

import (
	"bytes"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/utils"
	"path/filepath"
	"strings"
	"time"
)

func LintPaths(paths []string, strict bool) error {
	errors := make([]error, 0)
	fileCnt := 0
	okCnt := 0

	for _, path := range paths {
		snippetFiles, err := listSnippets(path)
		if err != nil {
			return err
		}

		for _, file := range snippetFiles {
			fileCnt += 1

			logger.Debug("Linting %s", file)
			snippet, err := parseSnippet(file, "https://github.com/Graylog2/___linter-test___.git")
			if err != nil {
				errors = append(errors, fmt.Errorf("linter error in file %s: %w", file, err))
				continue
			}

			if strings.TrimSpace(snippet.Message) == "" {
				errors = append(errors, fmt.Errorf("linter error in file %s: message cannot be empty", file))
				continue
			}

			urlList := make([]string, 0)
			for _, issuesOrPulls := range [][]string{snippet.Issues, snippet.PullRequests} {
				for _, value := range issuesOrPulls {
					if strings.TrimSpace(value) == "" {
						continue
					}
					url, err := utils.ResolveGitHubIssueURL(snippet.GitHubRepoURL, value)
					if err != nil {
						errors = append(errors, fmt.Errorf("linter error in file %s: %w", file, err))
						continue
					}
					urlList = append(urlList, url)
				}
			}

			// The GitHub security advisories (GHSA) might not have linked issues or pulls because the PRs might be
			// created in private forks as part of the security advisory process.
			if len(urlList) == 0 && !strings.HasPrefix(strings.ToLower(filepath.Base(file)), "ghsa-") {
				errors = append(errors, fmt.Errorf("linter error in file %s: at least one issue or pull request number needs to be present", file))
				continue
			}

			logger.Debug("Rendering %s", file)
			renderConfig := Config{
				RenderFormat:            FormatMD,
				RenderGitHubLinks:       true,
				SnippetsPaths:           []string{file},
				ReleaseDate:             time.Now().Format(time.DateOnly),
				ReleaseVersion:          "1.0.0",
				Product:                 "Render Test",
				SkipHeader:              false,
				RenderNoChanges:         false,
				SkipInvalidSnippets:     false,
				ReadStdin:               false,
				MarkdownHeaderBaseLevel: 1,
				GitHubRepoURL:           "https://github.com/Graylog2/___linter-test___.git",
			}
			var output bytes.Buffer
			if err := Render(renderConfig, &output); err != nil {
				errors = append(errors, fmt.Errorf("linter error in file: %s: %w", file, err))
			}
			logger.Debug("Render output for file %s:\n%s", file, output.String())

			okCnt += 1
		}
	}

	logger.Info("Linted %d snippet file(s) - ok=%d error=%d", fileCnt, okCnt, len(errors))

	if len(errors) > 0 {
		for _, e := range errors {
			logger.Error(e.Error())
		}

		return fmt.Errorf("detected errors in %d file(s)", len(errors))
	}

	if okCnt == 0 && len(errors) == 0 && strict {
		return fmt.Errorf("no files found for path(s): %s", strings.Join(paths, ", "))
	}

	return nil
}
