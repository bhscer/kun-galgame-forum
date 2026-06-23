package repository

import (
	"strings"
	"testing"
	"time"
)

// stripParens removes every parenthesized group so only top-level SQL remains —
// a subquery's WHERE (e.g. a correlated STRING_AGG, or an EXISTS) must not be
// mistaken for the base table's WHERE.
func stripParens(s string) string {
	var b strings.Builder
	depth := 0
	for _, r := range s {
		switch r {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// TestSourcesBaseWhereFlagConsistent guards the failure mode that bit the keyset
// rewrite: a source whose Query already ends in a top-level WHERE but whose
// HasBaseWhere is unset makes sourceQuery emit `WHERE … WHERE …` once the cursor
// / resource / SFW predicate is appended. Flagging it wrongly is caught too.
func TestSourcesBaseWhereFlagConsistent(t *testing.T) {
	for typeStr, src := range Sources {
		hasTopWhere := strings.Contains(
			strings.ToUpper(stripParens(src.Query)), " WHERE ",
		)
		if hasTopWhere != src.HasBaseWhere {
			t.Errorf("%s: HasBaseWhere=%v, but a top-level WHERE in Query was detected=%v — they must match",
				typeStr, src.HasBaseWhere, hasTopWhere)
		}
	}
}

// TestBuildKeysetSQLPlaceholderParity ensures every `?` in the generated SQL has
// exactly one bound arg (a mismatch silently misbinds the cursor). Covered for
// the first page (nil cursor → no placeholders) and a cursor page, SFW on so the
// SFW predicate is appended too.
func TestBuildKeysetSQLPlaceholderParity(t *testing.T) {
	all := make([]ActivitySource, 0, len(Sources))
	for _, s := range Sources {
		all = append(all, s)
	}
	cases := []*Cursor{
		nil,
		{Created: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), TypeStr: "TOPIC_CREATION", ID: 1},
	}
	for _, cur := range cases {
		sql, args := buildKeysetSQL(all, 50, cur, true, false, "")
		if got := strings.Count(sql, "?"); got != len(args) {
			t.Errorf("cursor=%v: %d placeholders but %d args", cur != nil, got, len(args))
		}
	}
}

// TestBuildKeysetSQLPerBranchExactCut guards the 资源-feed freeze fix: each source
// sub-query must carry the EXACT total-order cut (down to t.id) plus an id
// tiebreaker in its ORDER BY. Without it, a cluster of equal-`created` rows fills
// the LIMIT window with rows the outer cut rejects, the branch returns 0, and the
// feed reports a premature end (single-source feeds like 资源 stop scrolling).
func TestBuildKeysetSQLPerBranchExactCut(t *testing.T) {
	one := []ActivitySource{Sources["GALGAME_RESOURCE_CREATION"]}
	cur := &Cursor{Created: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC), TypeStr: "GALGAME_RESOURCE_CREATION", ID: 42}
	sql, _ := buildKeysetSQL(one, 30, cur, false, false, "")
	if !strings.Contains(sql, "t.id <") {
		t.Errorf("per-branch keyset cut missing the t.id tiebreaker — equal-created clusters can freeze the feed:\n%s", sql)
	}
	if !strings.Contains(sql, "ORDER BY t.created DESC, t.id DESC") {
		t.Errorf("per-branch ORDER BY missing the id tiebreaker:\n%s", sql)
	}
}
