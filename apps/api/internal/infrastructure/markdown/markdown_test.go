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
