// Package render compiles Markdown to HTML using goldmark, with custom
// renderers for ```mermaid fences and `$…$` / `$$…$$` math, plus TOC
// extraction. Spec ref: §9.
package render

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// TOCEntry is one heading row in the per-page right-rail TOC.
type TOCEntry struct {
	ID    string
	Text  string
	Level int
}

// Renderer renders Markdown to HTML and extracts a TOC.
type Renderer struct {
	md goldmark.Markdown
}

// New returns a configured Renderer.
func New() *Renderer {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			extension.Strikethrough,
			extension.Table,
			extension.TaskList,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(), // raw HTML allowed — internal docs trust authors
		),
	)
	return &Renderer{md: md}
}

// Render returns HTML and a TOC. Body is the markdown body (already
// stripped of front-matter by the caller).
func (r *Renderer) Render(body []byte) ([]byte, []TOCEntry, error) {
	// Run goldmark to produce AST then HTML.
	reader := text.NewReader(body)
	doc := r.md.Parser().Parse(reader)
	toc := extractTOC(doc, body)

	// Pre-process mermaid + math BEFORE goldmark renders. Both inject
	// <pre class="mermaid"> / <span class="math"> placeholders that
	// goldmark's HTML renderer passes through (WithUnsafe is set);
	// client-side scripts in Task 9 hydrate them.
	processed := preprocessMermaid(body)
	processed = preprocessMath(processed)
	var buf bytes.Buffer
	if err := r.md.Convert(processed, &buf); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), toc, nil
}

func extractTOC(doc ast.Node, src []byte) []TOCEntry {
	var out []TOCEntry
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		var textBuf bytes.Buffer
		ast.Walk(h, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}
			if tn, ok := child.(*ast.Text); ok {
				textBuf.Write(tn.Segment.Value(src))
			}
			return ast.WalkContinue, nil
		})
		id, _ := h.AttributeString("id")
		idStr := ""
		switch v := id.(type) {
		case []byte:
			idStr = string(v)
		case string:
			idStr = v
		}
		out = append(out, TOCEntry{
			ID:    idStr,
			Text:  textBuf.String(),
			Level: h.Level,
		})
		return ast.WalkContinue, nil
	})
	return out
}

var (
	mermaidRe  = regexp.MustCompile("(?s)```mermaid\\s*\n(.*?)\n```[ \\t]*(?:\n|$)")
	blockMath  = regexp.MustCompile(`(?s)\$\$(.+?)\$\$`)
	inlineMath = regexp.MustCompile(`\$([^\$\n\s][^\$\n]*?[^\$\n\s]|\S)\$`)
)

func preprocessMermaid(src []byte) []byte {
	return mermaidRe.ReplaceAllFunc(src, func(match []byte) []byte {
		m := mermaidRe.FindSubmatch(match)
		if len(m) < 2 {
			return match
		}
		var buf bytes.Buffer
		buf.WriteString(`<pre class="mermaid">`)
		buf.Write(util.EscapeHTML(m[1]))
		buf.WriteString(`</pre>`)
		return buf.Bytes()
	})
}

// preprocessMath wraps `$$...$$` and `$...$` in <span class="math"> so
// client-side KaTeX (Task 9) can hydrate them. Block-vs-inline distinction
// is deferred (YAGNI) — the class is "math" only. If a future task needs
// the distinction, add a `data-display` attribute.
func preprocessMath(src []byte) []byte {
	src = blockMath.ReplaceAllFunc(src, func(match []byte) []byte {
		inner := blockMath.FindSubmatch(match)[1]
		var buf bytes.Buffer
		buf.WriteString(`<span class="math">`)
		buf.Write(util.EscapeHTML(inner))
		buf.WriteString(`</span>`)
		return buf.Bytes()
	})
	src = inlineMath.ReplaceAllFunc(src, func(match []byte) []byte {
		inner := inlineMath.FindSubmatch(match)[1]
		var buf bytes.Buffer
		buf.WriteString(`<span class="math">`)
		buf.Write(util.EscapeHTML(inner))
		buf.WriteString(`</span>`)
		return buf.Bytes()
	})
	return src
}
