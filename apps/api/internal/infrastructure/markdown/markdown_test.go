package markdown

import (
	"strings"
	"testing"
)

// goldmark runs with html.WithUnsafe(); these assert the bluemonday pass strips
// the dangerous markup that raw user HTML can smuggle through.
func TestRenderStripsXSS(t *testing.T) {
	cases := []struct {
		name   string
		md     string
		banned []string
	}{
		{"script", "hi <script>alert(1)</script> there", []string{"<script", "alert(1)"}},
		{"img onerror", "<img src=x onerror=alert(1)>", []string{"onerror"}},
		{"js link", "[click](javascript:alert(1))", []string{"javascript:"}},
		{"iframe", "<iframe src=//evil.com></iframe>", []string{"<iframe"}},
		{"onclick", `<p onclick="x()">p</p>`, []string{"onclick"}},
		{"style tag", "<style>*{x:y}</style>t", []string{"<style"}},
		{"svg onload", `<svg onload=alert(1)>`, []string{"onload"}},
		{"spoiler xss", "||<script>bad()</script>||", []string{"<script", "bad()"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := Render(c.md)
			for _, b := range c.banned {
				if strings.Contains(out, b) {
					t.Errorf("must NOT contain %q\n got: %s", b, out)
				}
			}
		})
	}
}

// Assert every feature the pipeline produces survives the sanitizer.
func TestRenderPreservesFeatures(t *testing.T) {
	cases := []struct {
		name string
		md   string
		want []string
	}{
		{"heading id", "## Hello World", []string{"<h2", "id="}},
		{"safe link", "[ok](https://example.com)", []string{"https://example.com"}},
		{"bold italic", "**b** _i_", []string{"<strong>b</strong>", "<em>i</em>"}},
		{"code block", "```go\nfmt.Println()\n```", []string{"kun-code-container", "language-go", "copy"}},
		{"code block no lang", "```\nplain text\n```", []string{"kun-code-container", "copy"}},
		{"table", "| a | b |\n|---|---|\n| 1 | 2 |", []string{"kun-table-container", "<table>"}},
		{"spoiler", "||secret||", []string{"kun-spoiler", "secret"}},
		{"video", "kv:[v](https://e.com/x.mp4)", []string{"<video", "e.com/x.mp4"}},
		{"lazy image", "![alt](https://e.com/i.png)", []string{"data-kun-lazy-image", "loading=", "e.com/i.png"}},
		{"list", "- one\n- two", []string{"<ul>", "<li>"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := Render(c.md)
			for _, w := range c.want {
				if !strings.Contains(out, w) {
					t.Errorf("must contain %q\n got: %s", w, out)
				}
			}
		})
	}
}

// Regression: a no-language fenced block used to get the `.kun-code-container`
// closing `</div>` without an opening one, leaving a STRAY `</div>` that the
// browser used to close an ancestor early — diverging the SSR DOM from the
// client and cascading hydration mismatches through the rest of the post.
// Every code block (lang or not) must be div-balanced.
func TestRenderCodeBlocksDivBalanced(t *testing.T) {
	cases := map[string]string{
		"no lang":         "before\n\n```\nplain\n```\n\nafter",
		"with lang":       "before\n\n```go\nx := 1\n```\n\nafter",
		"mixed":           "```\nplain\n```\n\ntext\n\n```sql\nselect 1\n```",
		"indented":        "    indented code\n",
		"angle-bracket <": "```\nif a < b then\n```",
	}
	for name, md := range cases {
		t.Run(name, func(t *testing.T) {
			out := Render(md)
			if open, clo := strings.Count(out, "<div"), strings.Count(out, "</div>"); open != clo {
				t.Errorf("unbalanced divs: %d open vs %d close\n got: %s", open, clo, out)
			}
		})
	}
}
