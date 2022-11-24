package changelog

import (
	"bytes"
	"fmt"
	"github.com/Graylog2/graylog-project-cli/utils"
	"github.com/samber/lo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

const FormatHTML = "html"
const FormatD360HTML = "d360html"
const FormatMarkdown = "markdown"
const FormatMD = "md"

var AvailableFormatters = []string{FormatHTML, FormatD360HTML, FormatMD, FormatMarkdown}

var renderers = map[string]Renderer{}

var titleCaser = cases.Title(language.English)

type Renderer interface {
	RenderHeader(config Config, buf *bytes.Buffer) error

	RenderType(config Config, snippetType string, buf *bytes.Buffer) error

	RenderSnippets(config Config, snippets []Snippet, buf *bytes.Buffer) error
}

type SnippetRenderError struct {
	snippet Snippet
	err     error
}

func (e SnippetRenderError) Error() string {
	return fmt.Errorf("error for snippet %s: %w", e.snippet.Filename, e.err).Error()
}

func iterateIssuesAndPulls(snippet Snippet, callback func(string, string) error) error {
	for _, issuesOrPulls := range [][]string{snippet.Issues, snippet.PullRequests} {
		for _, value := range issuesOrPulls {
			if strings.TrimSpace(value) == "" {
				continue
			}
			issueURL, err := utils.ResolveGitHubIssueURL(snippet.GitHubRepoURL, value)
			if err != nil {
				return &SnippetRenderError{snippet, err}
			}

			title := utils.PrettifyGitHubIssueURL(issueURL, utils.PrettyModeRepo)

			if err := callback(title, issueURL); err != nil {
				return &SnippetRenderError{snippet, err}
			}
		}
	}

	return nil
}

func mdFormatContributors(snippet Snippet) []string {
	if len(snippet.Contributors) == 0 {
		return nil
	}

	names := lo.Filter[string](snippet.Contributors, func(name string, _ int) bool {
		return strings.TrimSpace(name) != ""
	})

	if len(names) == 0 {
		return nil
	}

	return lo.Map[string, string](snippet.Contributors, func(name string, _ int) string {
		if strings.HasPrefix(strings.TrimSpace(name), "@") {
			nameWithoutPrefix := strings.TrimPrefix(name, "@")
			return fmt.Sprintf(`[%s](https://github.com/%s)`, name, nameWithoutPrefix)
		} else {
			return name
		}
	})
}

type HTMLFormatter struct {
}

func (h HTMLFormatter) RenderHeader(config Config, buf *bytes.Buffer) error {
	buf.WriteString(fmt.Sprintf("<h1>%s %s</h1>\n\n", config.Product, config.ReleaseVersion))
	buf.WriteString(fmt.Sprintf("<p>Released: %s</p>\n\n", config.ReleaseDate))
	return nil
}

func (h HTMLFormatter) RenderType(config Config, snippetType string, buf *bytes.Buffer) error {
	buf.WriteString("<h2>")
	buf.WriteString(titleCaser.String(snippetType))
	buf.WriteString("</h2>\n")
	return nil
}

func (h HTMLFormatter) RenderSnippets(config Config, snippets []Snippet, buf *bytes.Buffer) error {
	buf.WriteString("<ul>\n")
	for _, snippet := range snippets {
		buf.WriteString("  <li>")
		if err := mdMessageRenderer.Convert([]byte(snippet.Message), buf); err != nil {
			return fmt.Errorf("couldn't convert message to HTML \"%s\": %w", snippet.Message, err)
		}

		if config.RenderGitHubLinks {
			err := iterateIssuesAndPulls(snippet, func(title, url string) error {
				buf.WriteString(fmt.Sprintf(` <a href="%s">%s</a>`, url, title))
				return nil
			})
			if err != nil {
				return err
			}
		}

		contributors := mdFormatContributors(snippet)
		if len(contributors) > 0 {
			formattedContributors := make([]string, 0, len(contributors))
			for _, name := range contributors {
				var nameBuf bytes.Buffer
				if err := mdMessageRenderer.Convert([]byte(name), &nameBuf); err != nil {
					return fmt.Errorf("couldn't convert message to HTML \"%s\": %w", snippet.Message, err)
				}
				formattedContributors = append(formattedContributors, nameBuf.String())
			}

			buf.WriteString(fmt.Sprintf(` (Thanks: %s)`, strings.Join(formattedContributors, ", ")))
		}

		// TODO: Write details
		buf.WriteString("</li>\n")
	}
	buf.WriteString("</ul>\n")
	return nil
}

type D360HTMLFormatter struct {
	// This formatter is using our custom HTML formatting -- this can be deleted once we moved off of Document360
}

func (h D360HTMLFormatter) RenderHeader(config Config, buf *bytes.Buffer) error {
	buf.WriteString(fmt.Sprintf("<h2>%s %s</h2>\n\n", config.Product, config.ReleaseVersion))
	buf.WriteString(fmt.Sprintf("<p>Released: %s</p>\n\n", config.ReleaseDate))
	return nil
}

func (h D360HTMLFormatter) RenderType(config Config, snippetType string, buf *bytes.Buffer) error {
	buf.WriteString("<p><strong>")
	buf.WriteString(titleCaser.String(snippetType))
	buf.WriteString("</strong></p>\n")
	return nil
}

func (h D360HTMLFormatter) RenderSnippets(config Config, snippets []Snippet, buf *bytes.Buffer) error {
	buf.WriteString("<ul>\n")
	for _, snippet := range snippets {
		buf.WriteString("  <li>")
		if err := mdMessageRenderer.Convert([]byte(snippet.Message), buf); err != nil {
			return fmt.Errorf("couldn't convert message to HTML \"%s\": %w", snippet.Message, err)
		}

		if config.RenderGitHubLinks {
			err := iterateIssuesAndPulls(snippet, func(title, url string) error {
				buf.WriteString(fmt.Sprintf(` <a href="%s">%s</a>`, url, title))
				return nil
			})
			if err != nil {
				return err
			}
		}

		contributors := mdFormatContributors(snippet)
		if len(contributors) > 0 {
			formattedContributors := make([]string, 0, len(contributors))
			for _, name := range contributors {
				var nameBuf bytes.Buffer
				if err := mdMessageRenderer.Convert([]byte(name), &nameBuf); err != nil {
					return fmt.Errorf("couldn't convert message to HTML \"%s\": %w", snippet.Message, err)
				}
				formattedContributors = append(formattedContributors, nameBuf.String())
			}

			buf.WriteString(fmt.Sprintf(` (Thanks: %s)`, strings.Join(formattedContributors, ", ")))
		}

		// TODO: Write details
		buf.WriteString("</li>\n")
	}
	buf.WriteString("</ul>\n")
	return nil
}

type MarkdownFormatter struct {
}

func (m MarkdownFormatter) RenderHeader(config Config, buf *bytes.Buffer) error {
	buf.WriteString(fmt.Sprintf("# %s %s\n\n", config.Product, config.ReleaseVersion))
	buf.WriteString(fmt.Sprintf("Released: %s\n\n", config.ReleaseDate))
	return nil
}

func (m MarkdownFormatter) RenderType(config Config, snippetType string, buf *bytes.Buffer) error {
	buf.WriteString("## ")
	buf.WriteString(titleCaser.String(snippetType))
	buf.WriteString("\n\n")
	return nil
}

func (m MarkdownFormatter) RenderSnippets(config Config, snippets []Snippet, buf *bytes.Buffer) error {
	for _, snippet := range snippets {
		buf.WriteString("- ")
		buf.WriteString(snippet.Message)

		if config.RenderGitHubLinks {
			err := iterateIssuesAndPulls(snippet, func(title, url string) error {
				buf.WriteString(fmt.Sprintf(` [%s](%s)`, title, url))
				return nil
			})
			if err != nil {
				return err
			}
		}

		contributors := mdFormatContributors(snippet)
		if len(contributors) > 0 {
			buf.WriteString(fmt.Sprintf(` (Thanks: %s)`, strings.Join(contributors, ", ")))
		}

		buf.WriteString("\n")
	}
	return nil
}

func init() {
	renderers[FormatHTML] = HTMLFormatter{}
	renderers[FormatD360HTML] = D360HTMLFormatter{}
	renderers[FormatMD] = MarkdownFormatter{}
	renderers[FormatMarkdown] = MarkdownFormatter{}
}

func GetRenderer(format string) (Renderer, error) {
	if renderer, ok := renderers[strings.ToLower(format)]; ok {
		return renderer, nil
	}

	return nil, fmt.Errorf("couldn't find renderer for \"%s\"", format)
}
