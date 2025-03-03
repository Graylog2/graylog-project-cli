package gh

import (
	"bufio"
	"io"
	"regexp"
)

var pullRequestDependencyPattern = regexp.MustCompile(`^/(?:jenkins-pr-deps|jpd|prd)\s+((?:Graylog2/\S+?#|https?://github.com/Graylog2/\S+?/pull/)[0-9]+)`)

func ParsePullDependencies(input io.Reader) ([]string, error) {
	dependencies := make([]string, 0)

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		match := pullRequestDependencyPattern.FindStringSubmatch(scanner.Text())

		if len(match) < 2 {
			continue
		}

		dependencies = append(dependencies, match[1])
	}

	return dependencies, nil
}
