// Single source of truth for "which URL to use as a galgame's head image".
//
// Kungal's rewriteBanners walker already resolves hashes to URLs on the
// server side (see apps/api/internal/galgame/client/banner.go) — both
// the U2 `effective_banner_hash` → `effective_banner_url` injection and
// the transition-era `banner_image_hash` → `banner` fill. So the FE
// fallback collapses to two practical levels:
//
//   1. `effective_banner_url` — the U2 canonical head image (derived
//      from `covers[sort_order=0]`). Preferred whenever present.
//   2. `banner` — legacy URL OR a URL kungal already filled from
//      banner_image_hash during transition. Either way it's a usable
//      string here.
//
// Returning `''` means "no banner available" — callers can show a
// placeholder. We intentionally do NOT fall back to a hardcoded default
// URL: that's the caller's policy decision (some surfaces want a CDN
// placeholder, others want a Nuxt /placeholder.webp asset).
//
// K-PR6 will drop `banner_image_hash` and the legacy `banner` URL field;
// at that point this util becomes a one-liner. Keep the shape ready.
interface BannerSource {
  effective_banner_url?: string
  banner?: string
}

export const getEffectiveBanner = (g?: BannerSource | null): string => {
  if (!g) return ''
  const eff = g.effective_banner_url?.trim()
  if (eff) return eff
  return g.banner?.trim() ?? ''
}
