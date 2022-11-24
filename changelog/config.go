package changelog

type Config struct {
	RenderFormat      string
	RenderGitHubLinks bool
	SnippetsPaths     []string
	ReleaseDate       string
	ReleaseVersion    string
	Product           string
}
