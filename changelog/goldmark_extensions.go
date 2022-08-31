package changelog

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var noParagraphRendererExtension = &noParagraphRenderer{}

// Markdown renderer for single message lines. Register our custom extension to avoid rendering paragraphs for
// the single line message.
var mdMessageRenderer = goldmark.New(goldmark.WithExtensions(noParagraphRendererExtension))

// Markdown renderer for the details. The details can contain multiple lines.
var mdDetailsRenderer = goldmark.New(goldmark.WithExtensions(extension.GFM))

// A custom goldmark renderer to avoid rendering paragraphs ("<p>") for single line messages.
type noParagraphRenderer struct {
}

func (n *noParagraphRenderer) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(n, 500),
	))
}

func (n *noParagraphRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindParagraph, n.renderParagraph)
}
func (r *noParagraphRenderer) renderParagraph(w util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	return ast.WalkContinue, nil
}
