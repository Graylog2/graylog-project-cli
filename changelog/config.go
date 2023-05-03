package changelog

import "regexp"

var SemverVersionPattern = regexp.MustCompile("^\\d+\\.\\d+\\.\\d+$")
var SemverVersionPatternWithPreRelease = regexp.MustCompile("^\\d+\\.\\d+\\.\\d+(:?\\-(:?alpha|beta|rc)\\.\\d+)?$")

type Config struct {
	RenderFormat            string
	RenderGitHubLinks       bool
	SnippetsPaths           []string
	ReleaseDate             string
	ReleaseVersion          string
	Product                 string
	SkipHeader              bool
	SkipInvalidSnippets     bool
	ReadStdin               bool
	MarkdownHeaderBaseLevel int
}
