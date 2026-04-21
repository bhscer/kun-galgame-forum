package markdown

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

var (
	spoilerRegex   = regexp.MustCompile(`\|\|(.*?)\|\|`)
	videoLinkRegex = regexp.MustCompile(`kv:<a href="(https?://[^\s]+?\.(mp4))">[^<]+</a>`)
	codeBlockRegex = regexp.MustCompile(`(?s)<pre><code class="language-(\w+)"`)

	md goldmark.Markdown
)

// TocLink is one entry in a table of contents tree. The JSON shape matches
// DocTocLink on the frontend.
type TocLink struct {
	ID       string    `json:"id"`
	Text     string    `json:"text"`
	Depth    int       `json:"depth"`
	Children []TocLink `json:"children,omitempty"`
}

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			mathjax.MathJax,
			highlighting.NewHighlighting(
				highlighting.WithStyle("monokai"),
				highlighting.WithGuessLanguage(true),
			),
			&h1ToH2Extension{},
			&lazyImageExtension{},
		),
		// Heading ids: enable goldmark's auto-heading-id transformer; the
		// actual slug algorithm is injected per-call via parser.WithIDs so
		// Chinese/Japanese headings keep their Unicode characters instead
		// of being stripped to empty strings.
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
}

// Render converts markdown to HTML with all custom transformations.
func Render(source string) string {
	result, _ := renderWithOptionalTOC(source, false)
	return result
}

// RenderWithTOC converts markdown to HTML and also returns a nested TOC tree
// built from the document's h2/h3 headings (h1 is promoted to h2 to match the
// h1→h2 render transform).
func RenderWithTOC(source string) (string, []TocLink) {
	return renderWithOptionalTOC(source, true)
}

func renderWithOptionalTOC(source string, wantTOC bool) (string, []TocLink) {
	src := []byte(source)
	reader := text.NewReader(src)
	ctx := parser.NewContext(parser.WithIDs(newUnicodeIDs()))
	root := md.Parser().Parse(reader, parser.WithContext(ctx))

	var toc []TocLink
	if wantTOC && root != nil {
		toc = buildTOCTree(collectHeadings(root, src), 3)
	}

	var buf bytes.Buffer
	if err := md.Renderer().Render(&buf, src, root); err != nil {
		return source, nil
	}

	result := buf.String()

	// Code block wrapper:
	// <pre><code class="language-go"... → wrapped in div.kun-code-container
	result = codeBlockRegex.ReplaceAllStringFunc(result, func(match string) string {
		lang := codeBlockRegex.FindStringSubmatch(match)
		if len(lang) < 2 {
			return match
		}
		return `<div class="kun-code-container language-` + lang[1] + `">` +
			`<div class="kun-code-header">` +
			`<span class="lang">` + lang[1] + `</span>` +
			`<button class="copy" title="Copy code"></button>` +
			`</div>` +
			`<pre><code class="language-` + lang[1] + `"`
	})
	result = strings.ReplaceAll(result, "</code></pre>", "</code></pre></div>")

	// Table wrapper
	result = strings.ReplaceAll(result, "<table>", `<div class="kun-table-container"><table>`)
	result = strings.ReplaceAll(result, "</table>", `</table></div>`)

	// Spoiler: ||text|| → <span class="kun-spoiler ...">text</span>
	result = spoilerRegex.ReplaceAllString(result,
		`<span class="kun-spoiler text-transparent kun-spoiler-hidden">$1</span>`)

	// Video: kv:<a href="url.mp4">...</a> → <video>
	result = videoLinkRegex.ReplaceAllString(result,
		`<video controls loop playsinline width="100%" src="$1"></video>`)

	return result, toc
}

// ──────────────────────────────────────────
// Heading IDs + TOC extraction
// ──────────────────────────────────────────

// flatHeading is one heading collected during the AST walk before the TOC
// tree is built.
type flatHeading struct {
	ID    string
	Text  string
	Depth int // effective depth after h1→h2 promotion
}

// unicodeIDs is a goldmark parser.IDs implementation that keeps Unicode
// letters/digits (CJK friendly) instead of stripping them to empty strings
// the way goldmark's default ASCII-only generator does. A fresh instance is
// created per Parse so dedupe state does not leak across documents.
type unicodeIDs struct {
	used map[string]int
	anon int
}

func newUnicodeIDs() *unicodeIDs {
	return &unicodeIDs{used: map[string]int{}}
}

func (u *unicodeIDs) Generate(value []byte, _ ast.NodeKind) []byte {
	base := slugify(string(value))
	if base == "" {
		base = fmt.Sprintf("heading-%d", u.anon)
		u.anon++
	}
	id := base
	if n := u.used[base]; n > 0 {
		u.used[base] = n + 1
		id = fmt.Sprintf("%s-%d", base, n)
	} else {
		u.used[base] = 1
	}
	u.Put([]byte(id))
	return []byte(id)
}

func (u *unicodeIDs) Put(value []byte) {
	s := string(value)
	if _, ok := u.used[s]; !ok {
		u.used[s] = 1
	}
}

// collectHeadings walks the AST and returns flat heading entries using the
// `id` attribute goldmark already stamped via unicodeIDs. h1 is promoted to
// depth 2 so TOC nesting aligns with the h1→h2 render transform.
func collectHeadings(root ast.Node, source []byte) []flatHeading {
	var out []flatHeading
	ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		var id string
		if attr, found := h.Attribute([]byte("id")); found {
			if b, ok := attr.([]byte); ok {
				id = string(b)
			}
		}
		depth := h.Level
		if depth == 1 {
			depth = 2
		}
		out = append(out, flatHeading{
			ID:    id,
			Text:  headingText(h, source),
			Depth: depth,
		})
		return ast.WalkContinue, nil
	})
	return out
}

// headingText concatenates the text content of a heading node, matching the
// visible rendered text (minus inline formatting).
func headingText(h *ast.Heading, source []byte) string {
	var b strings.Builder
	for c := h.FirstChild(); c != nil; c = c.NextSibling() {
		collectInlineText(c, source, &b)
	}
	return strings.TrimSpace(b.String())
}

func collectInlineText(n ast.Node, source []byte, b *strings.Builder) {
	switch v := n.(type) {
	case *ast.Text:
		b.Write(v.Segment.Value(source))
	case *ast.String:
		b.Write(v.Value)
	case *ast.CodeSpan:
		for c := v.FirstChild(); c != nil; c = c.NextSibling() {
			collectInlineText(c, source, b)
		}
	default:
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			collectInlineText(c, source, b)
		}
	}
}

// slugify builds a URL-friendly id that preserves Unicode letters/digits so
// Chinese and Japanese headings get meaningful ids instead of being stripped
// to empty strings like goldmark's default ASCII-only generator does.
func slugify(s string) string {
	s = strings.TrimSpace(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r):
			if r < 128 {
				b.WriteRune(unicode.ToLower(r))
			} else {
				b.WriteRune(r)
			}
			prevDash = false
		case unicode.IsDigit(r):
			b.WriteRune(r)
			prevDash = false
		case unicode.IsSpace(r), r == '-', r == '_':
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := strings.TrimRight(b.String(), "-")
	runes := []rune(out)
	if len(runes) > 100 {
		out = string(runes[:100])
	}
	return out
}

// buildTOCTree converts a flat heading list into a nested tree. Matches the
// legacy Nitro behavior: h1 is treated as depth 2, and headings deeper than
// maxDepth are excluded.
func buildTOCTree(flat []flatHeading, maxDepth int) []TocLink {
	if len(flat) == 0 {
		return nil
	}
	filtered := make([]flatHeading, 0, len(flat))
	for _, h := range flat {
		if h.Depth >= 2 && h.Depth <= maxDepth {
			filtered = append(filtered, h)
		}
	}
	if len(filtered) == 0 {
		return nil
	}

	var roots []TocLink
	// Stack holds pointers into parent `Children` slices (or into roots for
	// the top level) so we can append nested entries in place.
	type frame struct {
		depth int
		list  *[]TocLink
	}
	stack := []frame{{depth: 1, list: &roots}}

	for _, h := range filtered {
		for len(stack) > 1 && h.Depth <= stack[len(stack)-1].depth {
			stack = stack[:len(stack)-1]
		}
		top := stack[len(stack)-1]
		*top.list = append(*top.list, TocLink{
			ID:    h.ID,
			Text:  h.Text,
			Depth: h.Depth,
		})
		// Push a frame pointing at the new entry's children slice so the
		// next deeper heading lands inside it.
		newEntry := &(*top.list)[len(*top.list)-1]
		stack = append(stack, frame{depth: h.Depth, list: &newEntry.Children})
	}
	return roots
}

// ToPlainText strips markdown syntax and returns plain text, truncated to maxLen runes.
func ToPlainText(source string, maxLen int) string {
	text := source
	text = regexp.MustCompile(`!\[.*?\]\(.*?\)`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`\[([^\]]*)\]\(.*?\)`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile("[#*_~>`|]").ReplaceAllString(text, "")
	text = regexp.MustCompile(`\n{2,}`).ReplaceAllString(text, "\n")
	text = strings.TrimSpace(text)

	runes := []rune(text)
	if len(runes) > maxLen {
		return string(runes[:maxLen])
	}
	return text
}

// ──────────────────────────────────────────
// Extension: H1 → H2
// ──────────────────────────────────────────

type h1ToH2Extension struct{}

func (e *h1ToH2Extension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&h1ToH2Renderer{}, 100),
	))
}

type h1ToH2Renderer struct{}

func (r *h1ToH2Renderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindHeading, r.renderHeading)
}

func (r *h1ToH2Renderer) renderHeading(
	w util.BufWriter, source []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	n := node.(*ast.Heading)
	level := n.Level
	if level == 1 {
		level = 2
	}
	tag := byte('0' + level)

	if entering {
		w.WriteString("<h")
		w.WriteByte(tag)
		if n.Attributes() != nil {
			for _, attr := range n.Attributes() {
				w.WriteByte(' ')
				w.Write(attr.Name)
				w.WriteString(`="`)
				w.Write(util.EscapeHTML(attr.Value.([]byte)))
				w.WriteByte('"')
			}
		}
		w.WriteByte('>')
	} else {
		w.WriteString("</h")
		w.WriteByte(tag)
		w.WriteString(">\n")
	}
	return ast.WalkContinue, nil
}

// ──────────────────────────────────────────
// Extension: Lazy Image
// ──────────────────────────────────────────

type lazyImageExtension struct{}

func (e *lazyImageExtension) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&lazyImageRenderer{}, 100),
	))
}

type lazyImageRenderer struct{}

func (r *lazyImageRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, r.renderImage)
}

func (r *lazyImageRenderer) renderImage(
	w util.BufWriter, source []byte, node ast.Node, entering bool,
) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)

	// Collect alt text from child text nodes
	var altBuf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			altBuf.Write(t.Value(source))
		}
	}

	w.WriteString(`<img src="`)
	w.Write(util.EscapeHTML(n.Destination))
	w.WriteString(`" alt="`)
	w.Write(util.EscapeHTML(altBuf.Bytes()))
	if n.Title != nil {
		w.WriteString(`" title="`)
		w.Write(util.EscapeHTML(n.Title))
	}
	w.WriteString(`" loading="lazy" decoding="async" data-kun-lazy-image="true" />`)
	return ast.WalkSkipChildren, nil
}
