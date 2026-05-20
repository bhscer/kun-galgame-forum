package client

import (
	"bytes"
	"encoding/json"
	"strings"
)

// Banner resolution — server-side hash → CDN URL.
//
// Background: the Galgame Wiki is the SoT for galgame metadata. For
// banners uploaded through image_service (every NEW submission), wiki
// returns `banner: ""` + `banner_image_hash: "<sha256>"` — this is the
// documented, intended contract (docs/galgame_wiki/07-submission.md
// §banner; docs/image_service/06-integration-guide.md "新数据走 hash …
// 这是期望行为，不是 bug"). Legacy VNDB-synced rows keep a populated
// `banner` URL and an empty hash.
//
// The whole kungal frontend (cards, detail hero, /edit/galgame/mine
// list, SEO og:image) reads `galgame.banner` as a ready-to-use URL and
// has no hash→URL resolver. Rather than thread a resolver through every
// component + DTO, we resolve once here, at the single point every wiki
// payload funnels through (doRequest). Downstream — typed mappers AND
// verbatim json.RawMessage passthroughs like /galgame/mine — then see a
// plain usable `banner` with zero further changes.
//
// URL layout mirrors the wiki's imageclient.MainURL exactly
// (kun-oauth-admin/apps/api/pkg/imageclient/client.go):
//
//	{cdnBase}/{hash[:2]}/{hash[2:4]}/{hash}.webp
//
// cdnBase MUST equal the wiki's KUN_IMAGE_PUBLIC_BASE_URL.

// bannerURLFromHash builds the main-image CDN URL for a content hash.
// Returns "" when the hash is too short to fan out (mirrors the wiki
// helper's len < 4 guard) so callers can fall back safely.
func bannerURLFromHash(cdnBase, hash string) string {
	if len(hash) < 4 {
		return ""
	}
	return strings.TrimRight(cdnBase, "/") + "/" +
		hash[:2] + "/" + hash[2:4] + "/" + hash + ".webp"
}

// rewriteBanners walks the wiki response JSON and, for every object that
// carries an EMPTY `banner` alongside a non-empty `banner_image_hash`,
// fills `banner` with the resolved CDN URL. Objects whose `banner` is
// already a non-empty (legacy) URL are left untouched — this fixes the
// broken new-submission case without perturbing working legacy rows.
//
// Safety: any parse/encode failure returns the original bytes verbatim,
// so a malformed or unexpected payload can never be turned into an error
// by this cosmetic step. Numbers are decoded via json.Number so large
// IDs / timestamps round-trip with exact representation.
func rewriteBanners(raw json.RawMessage, cdnBase string) json.RawMessage {
	if cdnBase == "" || len(raw) == 0 {
		return raw
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var tree any
	if err := dec.Decode(&tree); err != nil {
		return raw
	}

	if !walkResolveBanner(tree, cdnBase) {
		// Nothing rewritten — avoid a needless re-marshal (and any
		// incidental key reordering) when there was no work to do.
		return raw
	}

	out, err := json.Marshal(tree)
	if err != nil {
		return raw
	}
	return out
}

// walkResolveBanner recurses through maps/slices, resolving banners in
// place. Returns true if it changed anything.
//
// Three injection rules — all defensive (only fill empties; never
// overwrite an existing legacy URL or a pre-resolved field):
//
//  1. (legacy U2-transition) Object with `banner_image_hash` non-empty
//     and empty `banner` → fill `banner` with the resolved URL. This
//     keeps existing FE code that reads `galgame.banner` working until
//     K-PR6 drops `banner_image_hash`.
//
//  2. (U2) Object with `effective_banner_hash` non-empty and no
//     `effective_banner_url` → fill `effective_banner_url`. This is the
//     forward-facing "head image" field (galgame detail / list cards).
//
//  3. (U2) Object with `image_hash` non-empty and no `cdn_url` → fill
//     `cdn_url`. Captures every row in a `covers[]` / `screenshots[]`
//     array — and any future image-bearing relation — without the
//     walker needing to know its parent key.
//
// Rules are independent (an object MAY trigger more than one). Walking
// proceeds into all children regardless, so nested revision/PR
// snapshots are resolved in the same pass.
func walkResolveBanner(node any, cdnBase string) bool {
	changed := false
	switch v := node.(type) {
	case map[string]any:
		// Rule 1: legacy banner_image_hash → banner URL (transition).
		if hashRaw, ok := v["banner_image_hash"]; ok {
			hash, _ := hashRaw.(string)
			bannerStr, _ := v["banner"].(string)
			if strings.TrimSpace(hash) != "" && strings.TrimSpace(bannerStr) == "" {
				if url := bannerURLFromHash(cdnBase, hash); url != "" {
					v["banner"] = url
					changed = true
				}
			}
		}
		// Rule 2: effective_banner_hash → effective_banner_url (U2).
		if hashRaw, ok := v["effective_banner_hash"]; ok {
			hash, _ := hashRaw.(string)
			existing, _ := v["effective_banner_url"].(string)
			if strings.TrimSpace(hash) != "" && strings.TrimSpace(existing) == "" {
				if url := bannerURLFromHash(cdnBase, hash); url != "" {
					v["effective_banner_url"] = url
					changed = true
				}
			}
		}
		// Rule 3: image_hash (covers / screenshots / future) → cdn_url.
		// Skipped when the object is the WIKI image-record itself (it has
		// its own `url` field) — the heuristic: cover/screenshot rows
		// also carry `sort_order`; image records do not. Cheap to guard.
		if hashRaw, ok := v["image_hash"]; ok {
			if _, isRow := v["sort_order"]; isRow {
				hash, _ := hashRaw.(string)
				existing, _ := v["cdn_url"].(string)
				if strings.TrimSpace(hash) != "" && strings.TrimSpace(existing) == "" {
					if url := bannerURLFromHash(cdnBase, hash); url != "" {
						v["cdn_url"] = url
						changed = true
					}
				}
			}
		}
		for _, child := range v {
			if walkResolveBanner(child, cdnBase) {
				changed = true
			}
		}
	case []any:
		for _, child := range v {
			if walkResolveBanner(child, cdnBase) {
				changed = true
			}
		}
	}
	return changed
}
