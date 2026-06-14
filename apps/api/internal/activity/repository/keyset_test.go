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
		sql, args := buildKeysetSQL(all, 50, cur, true, false)
		if got := strings.Count(sql, "?"); got != len(args) {
			t.Errorf("cursor=%v: %d placeholders but %d args", cur != nil, got, len(args))
		}
	}
}
