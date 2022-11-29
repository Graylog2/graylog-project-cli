package changelog

import (
	"fmt"
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/utils"
	"strings"
)

func LintPaths(paths []string) error {
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
			snippet, err := parseSnippet(file)
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

			if len(urlList) == 0 {
				errors = append(errors, fmt.Errorf("linter error in file %s: at least one issue or pull request number needs to be present", file))
				continue
			}

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

	return nil
}
