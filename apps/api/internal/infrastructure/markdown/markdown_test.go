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

// Mention: [@name](kungal-user:id) → a sanitizer-surviving link carrying the id
// (data-uid). The raw custom-scheme token must NOT leak (it would be a broken,
// scheme-stripped link otherwise).
func TestRenderMention(t *testing.T) {
	out := Render("hi [@白狐](kungal-user:123) there")
	for _, w := range []string{`class="kun-mention"`, `data-uid="123"`, "@白狐"} {
		if !strings.Contains(out, w) {
			t.Errorf("mention must contain %q\n got: %s", w, out)
		}
	}
	if strings.Contains(out, "kungal-user:") {
		t.Errorf("raw mention token leaked the custom scheme: %s", out)
	}
}

// Quote: [#floor](kungal-reply:id) → a quote span the frontend hydrates into a
// card. Carries the target reply id + floor; raw token must not leak.
func TestRenderQuote(t *testing.T) {
	out := Render("see [#12](kungal-reply:456)")
	for _, w := range []string{`class="kun-quote"`, `data-reply-id="456"`, `data-floor="12"`, "#12"} {
		if !strings.Contains(out, w) {
			t.Errorf("quote must contain %q\n got: %s", w, out)
		}
	}
	if strings.Contains(out, "kungal-reply:") {
		t.Errorf("raw quote token leaked the custom scheme: %s", out)
	}
}

// With a site base set, the mention href is absolute (UGCPolicy strips relative
// URLs, so a relative /user/N href would be dropped).
func TestRenderMentionAbsoluteHref(t *testing.T) {
	SetContentSiteBase("https://www.kungal.com")
	defer SetContentSiteBase("")
	out := Render("[@白狐](kungal-user:123)")
	if !strings.Contains(out, `href="https://www.kungal.com/user/123"`) {
		t.Errorf("mention should have an absolute href\n got: %s", out)
	}
}
