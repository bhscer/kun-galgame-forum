package client

import "testing"

func TestToWikiPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		// U3 taxonomy revisions / revert — keep under /galgame/<entity>/
		// namespace (decision §10 #1). MUST match before the generic
		// /galgame-tag → /tag rewrite.
		{"tag revisions list", "/api/galgame-tag/5/revisions", "/galgame/tag/5/revisions"},
		{"tag revision single", "/api/galgame-tag/5/revisions/3", "/galgame/tag/5/revisions/3"},
		{"tag revert", "/api/galgame-tag/5/revert", "/galgame/tag/5/revert"},
		{"official revisions list", "/api/galgame-official/9/revisions", "/galgame/official/9/revisions"},
		{"engine revert", "/api/galgame-engine/2/revert", "/galgame/engine/2/revert"},
		{"series revision single", "/api/galgame-series/7/revisions/1", "/galgame/series/7/revisions/1"},
		{"taxonomy /revisions/:rev/diff (defensive)", "/api/galgame-tag/5/revisions/3/diff", "/galgame/tag/5/revisions/3/diff"},

		// Regression: existing top-level endpoints on the SAME entities
		// must keep the bare single-word path (we only namespace revs).
		{"tag search untouched", "/api/galgame-tag/search", "/tag/search"},
		{"tag detail untouched", "/api/galgame-tag/5", "/tag/5"},
		{"tag PUT shape untouched", "/api/galgame-tag", "/tag"},
		{"official detail untouched", "/api/galgame-official/9", "/official/9"},
		{"series modal untouched", "/api/galgame-series/modal", "/series/modal"},
		{"resource untouched", "/api/galgame-resource/3", "/resource/3"},

		// Regression: galgame's own revisions / revert paths are NOT under
		// the /galgame-<entity>/ prefix (different shape) and must keep
		// flowing through the existing as-is / suffix routes.
		{"galgame :gid revisions", "/api/galgame/8/revisions", "/galgame/8/revisions"},
		{"galgame :gid revert", "/api/galgame/8/revert", "/galgame/8/revert"},
		{"galgame :gid /history/all → /revisions", "/api/galgame/8/history/all", "/galgame/8/revisions"},

		// Negative: junk suffixes on taxonomy entities still fall through
		// the namespace rule (only revisions/revert keep the namespace).
		{"tag :id/links would not exist but verify rule scoped",
			"/api/galgame-tag/5/links", "/tag/5/links"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ToWikiPath(c.in)
			if got != c.want {
				t.Errorf("ToWikiPath(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
