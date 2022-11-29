package changelog

import "regexp"

var SemverVersionPattern = regexp.MustCompile("^\\d+\\.\\d+\\.\\d+$")

type Config struct {
	RenderFormat      string
	RenderGitHubLinks bool
	SnippetsPaths     []string
	ReleaseDate       string
	ReleaseVersion    string
	Product           string
}
