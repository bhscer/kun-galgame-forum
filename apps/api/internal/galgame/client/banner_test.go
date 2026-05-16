package client

import (
	"encoding/json"
	"testing"
)

const cdn = "https://image.kungal.iloveren.link"

// hash → expected main URL: {cdn}/{hh}/{hh}/{hash}.webp
func TestBannerURLFromHash(t *testing.T) {
	cases := []struct {
		hash, want string
	}{
		{"abcd1234ef", cdn + "/ab/cd/abcd1234ef.webp"},
		{"00ff00ff00ff", cdn + "/00/ff/00ff00ff00ff.webp"},
		{"abc", ""},  // < 4 chars → unusable, mirrors wiki guard
		{"", ""},     // empty
	}
	for _, c := range cases {
		if got := bannerURLFromHash(cdn, c.hash); got != c.want {
			t.Errorf("bannerURLFromHash(%q) = %q, want %q", c.hash, got, c.want)
		}
	}
	// trailing slash on base must not double up
	if got := bannerURLFromHash(cdn+"/", "abcd"); got != cdn+"/ab/cd/abcd.webp" {
		t.Errorf("trailing-slash base not trimmed: %q", got)
	}
}

func TestRewriteBanners_DetailEnvelope(t *testing.T) {
	in := json.RawMessage(`{"galgame":{"id":60744,"banner":"","banner_image_hash":"abcd1234ef","status":3}}`)
	out := rewriteBanners(in, cdn)

	var got struct {
		Galgame struct {
			ID     int    `json:"id"`
			Banner string `json:"banner"`
			Status int    `json:"status"`
		} `json:"galgame"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v (out=%s)", err, out)
	}
	if got.Galgame.Banner != cdn+"/ab/cd/abcd1234ef.webp" {
		t.Errorf("banner not resolved: %q", got.Galgame.Banner)
	}
	// numbers must round-trip exactly (json.Number, no float mangling)
	if got.Galgame.ID != 60744 || got.Galgame.Status != 3 {
		t.Errorf("number round-trip broke: id=%d status=%d", got.Galgame.ID, got.Galgame.Status)
	}
}

func TestRewriteBanners_MineListMixed(t *testing.T) {
	// item[0]: new submission (empty banner + hash) → resolve
	// item[1]: legacy (populated banner, no hash)    → keep verbatim
	// item[2]: legacy w/ BOTH banner + hash          → keep legacy (empty-only)
	in := json.RawMessage(`{"items":[` +
		`{"id":1,"banner":"","banner_image_hash":"deadbeef00"},` +
		`{"id":2,"banner":"https://old.example/x.webp","banner_image_hash":""},` +
		`{"id":3,"banner":"https://old.example/y.webp","banner_image_hash":"cafe1234"}` +
		`],"total":3}`)
	out := rewriteBanners(in, cdn)

	var got struct {
		Items []struct {
			ID     int    `json:"id"`
			Banner string `json:"banner"`
		} `json:"items"`
		Total int `json:"total"`
	}
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Items[0].Banner != cdn+"/de/ad/deadbeef00.webp" {
		t.Errorf("item0 not resolved: %q", got.Items[0].Banner)
	}
	if got.Items[1].Banner != "https://old.example/x.webp" {
		t.Errorf("item1 legacy banner mutated: %q", got.Items[1].Banner)
	}
	if got.Items[2].Banner != "https://old.example/y.webp" {
		t.Errorf("item2 legacy-with-hash should keep legacy: %q", got.Items[2].Banner)
	}
	if got.Total != 3 {
		t.Errorf("sibling field lost: total=%d", got.Total)
	}
}

func TestRewriteBanners_FailSafe(t *testing.T) {
	// empty cdn → passthrough untouched
	raw := json.RawMessage(`{"galgame":{"banner":"","banner_image_hash":"abcd1234"}}`)
	if string(rewriteBanners(raw, "")) != string(raw) {
		t.Error("empty cdnBase should passthrough verbatim")
	}
	// non-JSON → returned verbatim, never errors
	bad := json.RawMessage(`Cannot GET /api/galgame/mine`)
	if string(rewriteBanners(bad, cdn)) != string(bad) {
		t.Error("non-JSON should passthrough verbatim")
	}
	// nothing to change → original bytes (no needless re-marshal)
	noop := json.RawMessage(`{"galgame":{"banner":"https://x/y.webp"}}`)
	if string(rewriteBanners(noop, cdn)) != string(noop) {
		t.Error("no-op payload should return original bytes")
	}
}
