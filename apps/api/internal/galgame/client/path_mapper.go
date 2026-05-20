package client

import "strings"

// wikiPathPrefixes maps frontend resource prefixes (kebab-case, with the
// "galgame-" namespace) to the corresponding wiki path prefix (single word).
var wikiPathPrefixes = []string{
	"/galgame-tag", "/galgame-official",
	"/galgame-engine", "/galgame-series",
	"/galgame-resource",
}

// wikiSuffixMap maps frontend detail sub-route suffixes to their wiki equivalent.
//
//	/pr/all       → /prs
//	/link/all     → /links
//	/history/all  → /revisions
var wikiSuffixMap = map[string]string{
	"/pr/all":      "/prs",
	"/link/all":    "/links",
	"/history/all": "/revisions",
}

// taxonomyEntities is the set whose revisions / revert endpoints live
// UNDER the /galgame/ namespace on the wiki side (per U3 decision §10
// #1: avoids name collision with topic.tag etc., and lines up with the
// galgame revision endpoints). Existing top-level endpoints on these
// entities (/tag/search, PUT /tag, DELETE /tag/:id, …) keep the bare
// single-word path — only the revisions/revert suffix is namespaced.
var taxonomyEntities = []string{"tag", "official", "engine", "series"}

// isTaxonomyRevSuffix returns true when the trailing path looks like a
// taxonomy revision read or revert (i.e. anything after the entity's
// :id that the namespaced rule should keep under /galgame/). Currently:
//
//	/revisions            list
//	/revisions/<rev>      single snapshot
//	/revisions/<rev>/diff (defensive: if wiki adds this for taxonomies later)
//	/revert               POST revert
func isTaxonomyRevSuffix(s string) bool {
	switch {
	case s == "/revisions":
		return true
	case strings.HasPrefix(s, "/revisions/"):
		return true
	case s == "/revert":
		return true
	}
	return false
}

// ToWikiPath converts a gateway request path (e.g. "/api/galgame-tag/foo") to
// the corresponding wiki service path (e.g. "/tag/foo"). It strips the
// leading "/api", then matches rules in order of specificity:
//
//  1. Taxonomy revisions / revert (NEW, U3) — preserved under /galgame/
//     namespace: /galgame-tag/:id/revisions[/:rev] → /galgame/tag/:id/...
//     Must run BEFORE the prefix-translation step, otherwise the
//     existing /galgame-tag → /tag rewrite eats the namespace.
//
//  2. Resource prefix translation — /galgame-tag/foo → /tag/foo
//     for all other paths on the taxonomy / resource entities.
//
//  3. Suffix translation — /galgame/:gid/pr/all → /galgame/:gid/prs (etc.)
//
// Paths that don't match any rule are returned as-is (minus /api).
func ToWikiPath(path string) string {
	// Strip "/api" prefix.
	wp := path[4:]

	// Rule 1: taxonomy revisions / revert under /galgame/<entity>/...
	for _, ent := range taxonomyEntities {
		prefix := "/galgame-" + ent + "/"
		if !strings.HasPrefix(wp, prefix) {
			continue
		}
		rest := wp[len(prefix):] // ":id/<suffix>"
		slash := strings.Index(rest, "/")
		if slash <= 0 {
			break // no further segment, fall through to rule 2
		}
		suffix := rest[slash:]
		if isTaxonomyRevSuffix(suffix) {
			return "/galgame/" + ent + "/" + rest
		}
		break
	}

	// Rule 2: resource prefix translation: "/galgame-tag/..." → "/tag/..."
	for _, prefix := range wikiPathPrefixes {
		wiki := "/" + prefix[len("/galgame-"):]
		if strings.HasPrefix(wp, prefix) &&
			(len(wp) == len(prefix) || wp[len(prefix)] == '/') {
			return wiki + wp[len(prefix):]
		}
	}

	// Rule 3: suffix translation: "/galgame/:gid/pr/all" → "/galgame/:gid/prs"
	for suffix, replacement := range wikiSuffixMap {
		if strings.HasSuffix(wp, suffix) {
			return wp[:len(wp)-len(suffix)] + replacement
		}
	}

	return wp
}
