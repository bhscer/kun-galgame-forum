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
func walkResolveBanner(node any, cdnBase string) bool {
	changed := false
	switch v := node.(type) {
	case map[string]any:
		hashRaw, hasHash := v["banner_image_hash"]
		if hasHash {
			hash, _ := hashRaw.(string)
			bannerStr, _ := v["banner"].(string)
			// Resolve only when there's a usable hash and no usable
			// legacy URL already present.
			if strings.TrimSpace(hash) != "" && strings.TrimSpace(bannerStr) == "" {
				if url := bannerURLFromHash(cdnBase, hash); url != "" {
					v["banner"] = url
					changed = true
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
