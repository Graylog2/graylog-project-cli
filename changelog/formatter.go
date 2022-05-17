package changelog

import (
	"bytes"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

const FormatHTML = "html"
const FormatMarkdown = "markdown"
const FormatMD = "md"

var AvailableFormatters = []string{FormatHTML, FormatMD, FormatMarkdown}

var renderers = map[string]Renderer{}

var titleCaser = cases.Title(language.English)

type Renderer interface {
	RenderType(snippetType string, buf *bytes.Buffer) error

	RenderSnippets(snippets []Snippet, buf *bytes.Buffer) error
}

type HTMLFormatter struct {
}

func (h HTMLFormatter) RenderType(snippetType string, buf *bytes.Buffer) error {
	buf.WriteString("<h2>")
	buf.WriteString(titleCaser.String(snippetType))
	buf.WriteString("</h2>\n")
	return nil
}

func (h HTMLFormatter) RenderSnippets(snippets []Snippet, buf *bytes.Buffer) error {
	buf.WriteString("<ul>\n")
	for _, snippet := range snippets {
		buf.WriteString("  <li>")
		if err := mdMessageRenderer.Convert([]byte(snippet.Message), buf); err != nil {
			return fmt.Errorf("couldn't convert message to HTML \"%s\": %w", snippet.Message, err)
		}
		// TODO: Write details
		buf.WriteString("</li>\n")
	}
	buf.WriteString("</ul>\n")
	return nil
}

type MarkdownFormatter struct {
}

func (m MarkdownFormatter) RenderType(snippetType string, buf *bytes.Buffer) error {
	buf.WriteString("## ")
	buf.WriteString(titleCaser.String(snippetType))
	buf.WriteString("\n\n")
	return nil
}

func (m MarkdownFormatter) RenderSnippets(snippets []Snippet, buf *bytes.Buffer) error {
	for _, snippet := range snippets {
		buf.WriteString("- ")
		buf.WriteString(snippet.Message)
		buf.WriteString("\n")
	}
	return nil
}

func init() {
	renderers[FormatHTML] = HTMLFormatter{}
	renderers[FormatMD] = MarkdownFormatter{}
	renderers[FormatMarkdown] = MarkdownFormatter{}
}

func GetRenderer(format string) (Renderer, error) {
	if renderer, ok := renderers[strings.ToLower(format)]; ok {
		return renderer, nil
	}

	return nil, fmt.Errorf("couldn't find renderer for \"%s\"", format)
}
